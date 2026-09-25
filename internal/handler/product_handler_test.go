package handler

import (
	"bytes"
	"context"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"testing"

	"stockopname-rita-backend/internal/dto"
)

func newProductHandler(productSvc *fakeProductService, importSvc *fakeImportService) ProductHandler {
	if importSvc == nil {
		importSvc = &fakeImportService{
			importFunc: func(ctx context.Context, produkData []byte, barcodeData []byte, dryRun bool) (dto.ImportResultResponse, error) {
				return dto.ImportResultResponse{}, nil
			},
		}
	}
	return NewProductHandler(productSvc, importSvc)
}

func TestCreateProduct(t *testing.T) {
	h := newProductHandler(&fakeProductService{
		createFunc: func(ctx context.Context, req dto.ProductCreateRequest) (dto.ProductResponse, error) {
			if req.PLU != "100251" || req.Name != "Sari Roti" || req.DepartmentCode != "1138" {
				t.Errorf("unexpected request %+v", req)
			}
			if req.BuyPrice == nil || *req.BuyPrice != 14500 {
				t.Errorf("unexpected buy price %+v", req)
			}
			return dto.ProductResponse{Id: 11, PLU: "100251"}, nil
		},
	}, nil)

	req := newJSONRequest(t, http.MethodPost, "/products", `{"plu":"100251","name":"Sari Roti","department_code":"1138","buy_price":14500,"sell_price":17200,"barcodes":["1002515550011"]}`)
	rec := httptest.NewRecorder()
	h.CreateProduct(rec, req, nil)

	if rec.Code != http.StatusCreated {
		t.Fatalf("expected status %d, got %d", http.StatusCreated, rec.Code)
	}
	if msg := decodeMessage(t, rec); msg != "Product created successfully" {
		t.Errorf("unexpected message %q", msg)
	}
}

func TestDeleteProduct(t *testing.T) {
	called := false
	h := newProductHandler(&fakeProductService{
		deleteFunc: func(ctx context.Context, id int) error {
			called = true
			if id != 11 {
				t.Errorf("expected id 11, got %d", id)
			}
			return nil
		},
	}, nil)

	req := newJSONRequest(t, http.MethodDelete, "/products/11", "")
	rec := httptest.NewRecorder()
	h.DeleteProduct(rec, req, paramsWithID("11"))

	if rec.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d", http.StatusOK, rec.Code)
	}
	if !called {
		t.Error("service Delete must be called")
	}
}

func TestGetProducts(t *testing.T) {
	h := newProductHandler(&fakeProductService{
		findAllFunc: func(ctx context.Context, pagination *dto.Pagination, search string) ([]dto.ProductResponse, error) {
			if pagination.Page != 1 || pagination.Limit != 10 || search != "roti" {
				t.Errorf("unexpected pagination/search %+v %q", pagination, search)
			}
			return []dto.ProductResponse{{Id: 11, PLU: "100251"}}, nil
		},
	}, nil)

	req := newJSONRequest(t, http.MethodGet, "/products?page=1&limit=10&search=roti", "")
	rec := httptest.NewRecorder()
	h.GetProducts(rec, req, nil)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d", http.StatusOK, rec.Code)
	}
}

func newImportRequest(t *testing.T, target string, produk []byte, barcode []byte) *http.Request {
	t.Helper()
	var body bytes.Buffer
	writer := multipart.NewWriter(&body)
	if produk != nil {
		part, err := writer.CreateFormFile("produk", "PRODUK.DBF")
		if err != nil {
			t.Fatalf("form file: %v", err)
		}
		if _, err := part.Write(produk); err != nil {
			t.Fatalf("write part: %v", err)
		}
	}
	if barcode != nil {
		part, err := writer.CreateFormFile("barcode", "BARCODE.DBF")
		if err != nil {
			t.Fatalf("form file: %v", err)
		}
		if _, err := part.Write(barcode); err != nil {
			t.Fatalf("write part: %v", err)
		}
	}
	if err := writer.Close(); err != nil {
		t.Fatalf("close writer: %v", err)
	}
	req := httptest.NewRequest(http.MethodPost, target, &body)
	req.Header.Set("Content-Type", writer.FormDataContentType())
	return req
}

func TestImportProducts(t *testing.T) {
	var gotProduk, gotBarcode []byte
	var gotDryRun bool
	h := newProductHandler(&fakeProductService{}, &fakeImportService{
		importFunc: func(ctx context.Context, produkData []byte, barcodeData []byte, dryRun bool) (dto.ImportResultResponse, error) {
			gotProduk, gotBarcode, gotDryRun = produkData, barcodeData, dryRun
			return dto.ImportResultResponse{ProductsInserted: 2, DryRun: dryRun}, nil
		},
	})

	req := newImportRequest(t, "/products/import?dry_run=true", []byte("produk-bytes"), []byte("barcode-bytes"))
	rec := httptest.NewRecorder()
	h.ImportProducts(rec, req, nil)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d", http.StatusOK, rec.Code)
	}
	if string(gotProduk) != "produk-bytes" || string(gotBarcode) != "barcode-bytes" || !gotDryRun {
		t.Errorf("service got wrong args produk=%q barcode=%q dryRun=%v", gotProduk, gotBarcode, gotDryRun)
	}
	if msg := decodeMessage(t, rec); msg != "Master data validation completed (dry run, nothing written)" {
		t.Errorf("unexpected message %q", msg)
	}
}

func TestImportProductsMissingFile(t *testing.T) {
	h := newProductHandler(&fakeProductService{}, nil)

	req := newImportRequest(t, "/products/import", []byte("produk-bytes"), nil)
	rec := httptest.NewRecorder()
	h.ImportProducts(rec, req, nil)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected status %d, got %d", http.StatusBadRequest, rec.Code)
	}
}

func TestClearProductsRequiresConfirm(t *testing.T) {
	called := false
	h := newProductHandler(&fakeProductService{
		clearFunc: func(ctx context.Context) error {
			called = true
			return nil
		},
	}, nil)

	req := newJSONRequest(t, http.MethodPost, "/products/clear", "")
	rec := httptest.NewRecorder()
	h.ClearProducts(rec, req, nil)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected status %d, got %d", http.StatusBadRequest, rec.Code)
	}
	if called {
		t.Error("service Clear must not be called without confirm")
	}
}

func TestClearProducts(t *testing.T) {
	called := false
	h := newProductHandler(&fakeProductService{
		clearFunc: func(ctx context.Context) error {
			called = true
			return nil
		},
	}, nil)

	req := newJSONRequest(t, http.MethodPost, "/products/clear?confirm=true", "")
	rec := httptest.NewRecorder()
	h.ClearProducts(rec, req, nil)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d", http.StatusOK, rec.Code)
	}
	if !called {
		t.Error("service Clear must be called")
	}
	if msg := decodeMessage(t, rec); msg != "Master data cleared successfully" {
		t.Errorf("unexpected message %q", msg)
	}
}
