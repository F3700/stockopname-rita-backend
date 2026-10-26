-- ============================================
-- SCHEMA: Stock Opname Rita (master produk per sesi)
--
-- sesi         : satu tempat/sesi stock opname
-- product      : master milik sesi; PLU unik per sesi
-- barcode      : 1 produk dapat memiliki N barcode; kode unik per sesi
-- stock_opname : histori immutable, memakai SNAPSHOT produk
--                (tanpa FK ke product agar tetap terbaca
--                walau master berubah / dihapus / di-clear)
-- ============================================

CREATE TABLE sesi (
    sesi_id INTEGER GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    sesi_location VARCHAR(100) NOT NULL,
    sesi_code VARCHAR(20) NOT NULL UNIQUE,
    sesi_status VARCHAR(20) NOT NULL
        CHECK (sesi_status IN (
            'IN_PROGRESS',
            'COMPLETED',
            'CANCELLED'
        )),
    sesi_startedat TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    sesi_endedat TIMESTAMPTZ,

    CONSTRAINT chk_sesi_time
        CHECK (sesi_endedat IS NULL OR sesi_endedat >= sesi_startedat),

    CONSTRAINT chk_sesi_status_time
        CHECK (
            (sesi_status = 'IN_PROGRESS' AND sesi_endedat IS NULL)
            OR
            (sesi_status IN ('COMPLETED', 'CANCELLED') AND sesi_endedat IS NOT NULL)
        )
);


CREATE TABLE product (
    product_id INTEGER GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    product_sesi_id INTEGER NOT NULL,
    product_plu VARCHAR(10) NOT NULL,
    product_name VARCHAR(150) NOT NULL,
    product_department_code VARCHAR(10) NOT NULL,
    product_buyprice NUMERIC(15,2) NOT NULL CHECK (product_buyprice >= 0),
    product_sellprice NUMERIC(15,2) NOT NULL CHECK (product_sellprice >= 0),
    product_createdat TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    product_updatedat TIMESTAMPTZ NOT NULL DEFAULT NOW(),

    CONSTRAINT fk_product_sesi
        FOREIGN KEY (product_sesi_id)
        REFERENCES sesi(sesi_id)
        ON DELETE CASCADE,

    CONSTRAINT uq_product_sesi_plu
        UNIQUE (product_sesi_id, product_plu),

    -- Dipakai composite FK barcode.
    CONSTRAINT uq_product_sesi_product
        UNIQUE (product_sesi_id, product_id)
);

CREATE INDEX idx_product_name ON product(product_name);
CREATE INDEX idx_product_dept ON product(product_department_code);


CREATE TABLE barcode (
    barcode_id INTEGER GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    barcode_sesi_id INTEGER NOT NULL,
    barcode_product_id INTEGER NOT NULL,
    barcode_code VARCHAR(15) NOT NULL,
    barcode_createdat TIMESTAMPTZ NOT NULL DEFAULT NOW(),

    CONSTRAINT uq_barcode_sesi_code
        UNIQUE (barcode_sesi_id, barcode_code),

    CONSTRAINT fk_barcode_sesi_product
        FOREIGN KEY (barcode_sesi_id, barcode_product_id)
        REFERENCES product(product_sesi_id, product_id)
        ON DELETE CASCADE
);

CREATE INDEX idx_barcode_product ON barcode(barcode_product_id);


CREATE TABLE coordinator (
    coor_id INTEGER GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    coor_code VARCHAR(20) NOT NULL,
    coor_sesi_id INTEGER NOT NULL,
    coor_status VARCHAR(20) NOT NULL
        CHECK (coor_status IN (
            'IN_PROGRESS',
            'IN_REVIEW',
            'COMPLETED',
            'CANCELLED'
        )),

    CONSTRAINT fk_coordinator_sesi
        FOREIGN KEY (coor_sesi_id)
        REFERENCES sesi(sesi_id)
        ON DELETE CASCADE,

    CONSTRAINT uq_coordinator_sesi_code
        UNIQUE (coor_sesi_id, coor_code)
);


CREATE TABLE upload (
    upload_id INTEGER GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    upload_coor_id INTEGER NOT NULL,
    upload_createdat TIMESTAMPTZ NOT NULL DEFAULT NOW(),

    CONSTRAINT fk_upload_coordinator
        FOREIGN KEY (upload_coor_id)
        REFERENCES coordinator(coor_id)
        ON DELETE CASCADE
);

CREATE INDEX idx_upload_coor ON upload(upload_coor_id);


CREATE TABLE rak (
    rak_id INTEGER GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    rak_name VARCHAR(15) NOT NULL,
    rak_updatedat TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    rak_coor_id INTEGER NOT NULL,
    rak_upload_id INTEGER,
    -- Snapshot item saat submit. Dipakai PDF; tidak berubah oleh koreksi web.
    rak_upload_snapshot JSONB NOT NULL DEFAULT '[]'::jsonb,

    CONSTRAINT fk_rak_coordinator
        FOREIGN KEY (rak_coor_id)
        REFERENCES coordinator(coor_id)
        ON DELETE CASCADE,

    CONSTRAINT fk_rak_upload
        FOREIGN KEY (rak_upload_id)
        REFERENCES upload(upload_id)
        ON DELETE CASCADE,

    CONSTRAINT uq_rak_coor_name
        UNIQUE (rak_coor_id, rak_name)
);

CREATE INDEX idx_rak_upload ON rak(rak_upload_id);


CREATE TABLE stock_opname (
    stock_opname_id INTEGER GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    stock_opname_quantity INTEGER NOT NULL
        CHECK (stock_opname_quantity >= 0),
    stock_opname_updatedat TIMESTAMPTZ NOT NULL DEFAULT NOW(),

    -- Product snapshot: kolom biasa, BUKAN foreign key.
    so_product_plu VARCHAR(10) NOT NULL,
    so_product_name VARCHAR(150) NOT NULL,
    so_barcode VARCHAR(15) NOT NULL DEFAULT '',
    so_buyprice NUMERIC(15,2) NOT NULL CHECK (so_buyprice >= 0),
    so_sellprice NUMERIC(15,2) NOT NULL CHECK (so_sellprice >= 0),

    stock_opname_rak_id INTEGER NOT NULL,

    CONSTRAINT fk_stock_opname_rak
        FOREIGN KEY (stock_opname_rak_id)
        REFERENCES rak(rak_id)
        ON DELETE CASCADE,

    CONSTRAINT uq_stock_opname_rak_plu
        UNIQUE (
            stock_opname_rak_id,
            so_product_plu
        )
);


CREATE OR REPLACE FUNCTION sync_sesi_endedat()
RETURNS TRIGGER AS $$
BEGIN

    IF NEW.sesi_status = 'IN_PROGRESS' THEN
        NEW.sesi_endedat := NULL;

    ELSE
        NEW.sesi_endedat := NOW();

    END IF;

    RETURN NEW;
END;
$$ LANGUAGE plpgsql;


CREATE TRIGGER trg_sync_sesi_endedat
BEFORE UPDATE OF sesi_status ON sesi
FOR EACH ROW
EXECUTE FUNCTION sync_sesi_endedat();


CREATE OR REPLACE FUNCTION update_sesi_status(
    p_sesi_id INTEGER,
    p_status VARCHAR
)
RETURNS VOID
LANGUAGE plpgsql
AS $$
BEGIN
    UPDATE sesi
    SET sesi_status = p_status
    WHERE sesi_id = p_sesi_id;

    IF p_status = 'COMPLETED' THEN
        UPDATE coordinator
        SET coor_status = 'COMPLETED'
        WHERE coor_sesi_id = p_sesi_id
          AND coor_status IN ('IN_PROGRESS', 'IN_REVIEW');

    ELSIF p_status = 'CANCELLED' THEN
        UPDATE coordinator
        SET coor_status = 'CANCELLED'
        WHERE coor_sesi_id = p_sesi_id
          AND coor_status IN ('IN_PROGRESS', 'IN_REVIEW');
    END IF;
END;
$$;


CREATE OR REPLACE FUNCTION recalc_coordinator_status(p_coor_id INTEGER)
RETURNS VOID AS $$
BEGIN
    UPDATE coordinator c
    SET coor_status = 'IN_REVIEW'
    WHERE c.coor_id = p_coor_id
      AND c.coor_status = 'IN_PROGRESS'
      AND EXISTS (SELECT 1 FROM rak r WHERE r.rak_coor_id = c.coor_id)
      AND NOT EXISTS (
          SELECT 1
          FROM rak r
          WHERE r.rak_coor_id = c.coor_id
            AND NOT EXISTS (
                SELECT 1
                FROM stock_opname so
                WHERE so.stock_opname_rak_id = r.rak_id
            )
      );
END;
$$ LANGUAGE plpgsql;


CREATE OR REPLACE FUNCTION sync_coordinator_status()
RETURNS TRIGGER AS $$
BEGIN

    PERFORM recalc_coordinator_status((
        SELECT r.rak_coor_id
        FROM rak r
        WHERE r.rak_id = NEW.stock_opname_rak_id
    ));

    RETURN NEW;
END;
$$ LANGUAGE plpgsql;


CREATE TRIGGER trg_sync_coordinator_status
AFTER INSERT ON stock_opname
FOR EACH ROW
EXECUTE FUNCTION sync_coordinator_status();
