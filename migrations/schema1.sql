CREATE TABLE department (
    department_id INTEGER GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    department_code VARCHAR(10) NOT NULL UNIQUE,
    department_name VARCHAR(100) NOT NULL,
    department_desc TEXT
);


CREATE TABLE category (
    category_id INTEGER GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    category_name VARCHAR(100) NOT NULL,
    category_desc TEXT
);


CREATE TABLE product (
    product_id INTEGER GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    product_barcode VARCHAR(20) NOT NULL UNIQUE,
    product_name VARCHAR(150) NOT NULL,
    product_buyprice NUMERIC(15,2) NOT NULL CHECK (product_buyprice >= 0),
    product_sellprice NUMERIC(15,2) NOT NULL CHECK (product_sellprice >= 0),
    product_createdat TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    product_updatedat TIMESTAMPTZ NOT NULL DEFAULT NOW(),

    product_category_id INTEGER NOT NULL,
    product_department_id INTEGER NOT NULL,

    CONSTRAINT fk_product_category_id
        FOREIGN KEY (product_category_id)
        REFERENCES category(category_id)
        ON DELETE RESTRICT,

    CONSTRAINT fk_product_department_id
        FOREIGN KEY (product_department_id)
        REFERENCES department(department_id)
        ON DELETE RESTRICT
);

CREATE INDEX idx_product_updatedat ON product(product_updatedat);

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


CREATE TABLE coordinator (
    coor_id INTEGER GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    coor_code VARCHAR(10) NOT NULL,
    coor_sesi_id INTEGER NOT NULL,

    CONSTRAINT fk_coordinator_sesi
        FOREIGN KEY (coor_sesi_id)
        REFERENCES sesi(sesi_id)
        ON DELETE RESTRICT,

    CONSTRAINT uq_coordinator_sesi_code
        UNIQUE (coor_sesi_id, coor_code)
);


CREATE TABLE inspector (
    inspector_id INTEGER GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    inspector_code VARCHAR(10) NOT NULL,
    inspector_startedat TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    inspector_endedat TIMESTAMPTZ,
    inspector_coor_id INTEGER NOT NULL,

    CONSTRAINT fk_inspector_coordinator
        FOREIGN KEY (inspector_coor_id)
        REFERENCES coordinator(coor_id)
        ON DELETE RESTRICT,

    CONSTRAINT uq_inspector_coor_code
        UNIQUE (inspector_coor_id, inspector_code),
    
    CONSTRAINT chk_inspector_time
        CHECK (
            inspector_endedat IS NULL
            OR inspector_endedat >= inspector_startedat
        )
);


CREATE TABLE rak (
    rak_id INTEGER GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    rak_name VARCHAR(100) NOT NULL,
    rak_updatedat TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    rak_inspector_id INTEGER NOT NULL,

    CONSTRAINT fk_rak_inspector
        FOREIGN KEY (rak_inspector_id)
        REFERENCES inspector(inspector_id)
        ON DELETE RESTRICT,
    
    CONSTRAINT uq_rak_inspector_name
        UNIQUE (rak_inspector_id, rak_name)
);


CREATE TABLE stock_opname (
    stock_opname_id INTEGER GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    stock_opname_quantity INTEGER NOT NULL
        CHECK (stock_opname_quantity >= 0),
    stock_opname_updatedat TIMESTAMPTZ NOT NULL DEFAULT NOW(),

    stock_opname_product_id INTEGER NOT NULL,
    stock_opname_rak_id INTEGER NOT NULL,

    CONSTRAINT fk_stock_opname_product
        FOREIGN KEY (stock_opname_product_id)
        REFERENCES product(product_id)
        ON DELETE RESTRICT,

    CONSTRAINT fk_stock_opname_rak
        FOREIGN KEY (stock_opname_rak_id)
        REFERENCES rak(rak_id)
        ON DELETE RESTRICT,

    CONSTRAINT uq_stock_opname_rak_product
        UNIQUE (
            stock_opname_rak_id,
            stock_opname_product_id
        )
);

CREATE TABLE deleted_product (
    product_id INTEGER NOT NULL PRIMARY KEY,
    deleted_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);