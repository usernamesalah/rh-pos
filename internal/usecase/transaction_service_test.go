package usecase_test

import (
	"testing"

	"github.com/usernamesalah/rh-pos/internal/domain/entities"
	"github.com/usernamesalah/rh-pos/internal/usecase"
)

func TestResolveItemPrice_RegularProduct_UsesHargaJual(t *testing.T) {
	price := 50000.0
	product := &entities.Product{HargaJual: &price, IsDynamicPrice: false}
	got := usecase.ResolveItemPrice(product, nil)
	if got != 50000.0 {
		t.Fatalf("expected 50000.0, got %v", got)
	}
}

func TestResolveItemPrice_RegularProduct_NilHargaJual_ReturnsZero(t *testing.T) {
	product := &entities.Product{HargaJual: nil, IsDynamicPrice: false}
	got := usecase.ResolveItemPrice(product, nil)
	if got != 0.0 {
		t.Fatalf("expected 0.0, got %v", got)
	}
}

func TestResolveItemPrice_DynamicProduct_UsesRequestPrice(t *testing.T) {
	product := &entities.Product{IsDynamicPrice: true}
	reqPrice := 75000.0
	got := usecase.ResolveItemPrice(product, &reqPrice)
	if got != 75000.0 {
		t.Fatalf("expected 75000.0, got %v", got)
	}
}

func TestResolveItemPrice_DynamicProduct_NilPrice_ReturnsZero(t *testing.T) {
	product := &entities.Product{IsDynamicPrice: true}
	got := usecase.ResolveItemPrice(product, nil)
	if got != 0.0 {
		t.Fatalf("expected 0.0, got %v", got)
	}
}

func TestResolveItemPrice_DynamicProduct_IgnoresHargaJual(t *testing.T) {
	storedPrice := 99999.0
	product := &entities.Product{HargaJual: &storedPrice, IsDynamicPrice: true}
	reqPrice := 25000.0
	got := usecase.ResolveItemPrice(product, &reqPrice)
	if got != 25000.0 {
		t.Fatalf("expected request price 25000.0, got %v", got)
	}
}
