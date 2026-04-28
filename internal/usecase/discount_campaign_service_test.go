package usecase

import (
	"testing"

	"github.com/usernamesalah/rh-pos/internal/domain/entities"
)

func TestValidateCampaignType_ProductPercentage_Valid(t *testing.T) {
	err := validateCampaignType(entities.CampaignTypeProductPercentageDiscount, 20, nil, nil, nil, nil)
	if err != nil {
		t.Errorf("expected no error, got: %v", err)
	}
}

func TestValidateCampaignType_ProductPercentage_InvalidPct(t *testing.T) {
	err := validateCampaignType(entities.CampaignTypeProductPercentageDiscount, 0, nil, nil, nil, nil)
	if err == nil {
		t.Error("expected error for 0 percentage")
	}
	err = validateCampaignType(entities.CampaignTypeProductPercentageDiscount, 101, nil, nil, nil, nil)
	if err == nil {
		t.Error("expected error for >100 percentage")
	}
}

func TestValidateCampaignType_BuyXAmount_Valid(t *testing.T) {
	err := validateCampaignType(entities.CampaignTypeBuyXQtyGetDiscountAmount, 0, intPtr(3), floatPtr(10000), nil, nil)
	if err != nil {
		t.Errorf("expected no error, got: %v", err)
	}
}

func TestValidateCampaignType_BuyXAmount_MissingBuyQty(t *testing.T) {
	err := validateCampaignType(entities.CampaignTypeBuyXQtyGetDiscountAmount, 0, nil, floatPtr(10000), nil, nil)
	if err == nil {
		t.Error("expected error for missing buy_quantity")
	}
}

func TestValidateCampaignType_BuyXAmount_LowBuyQty(t *testing.T) {
	err := validateCampaignType(entities.CampaignTypeBuyXQtyGetDiscountAmount, 0, intPtr(1), floatPtr(10000), nil, nil)
	if err == nil {
		t.Error("expected error for buy_quantity < 2")
	}
}

func TestValidateCampaignType_BuyXAmount_MissingDiscountAmount(t *testing.T) {
	err := validateCampaignType(entities.CampaignTypeBuyXQtyGetDiscountAmount, 0, intPtr(3), nil, nil, nil)
	if err == nil {
		t.Error("expected error for missing discount_amount")
	}
}

func TestValidateCampaignType_BuyXPercent_Valid(t *testing.T) {
	err := validateCampaignType(entities.CampaignTypeBuyXQtyGetDiscountPercent, 15, intPtr(3), nil, nil, nil)
	if err != nil {
		t.Errorf("expected no error, got: %v", err)
	}
}

func TestValidateCampaignType_BuyXPercent_BadPct(t *testing.T) {
	err := validateCampaignType(entities.CampaignTypeBuyXQtyGetDiscountPercent, 0, intPtr(3), nil, nil, nil)
	if err == nil {
		t.Error("expected error for 0 percentage")
	}
}

func TestValidateCampaignType_FreeItem_Valid(t *testing.T) {
	err := validateCampaignType(entities.CampaignTypeBuyXProductGetYProductFree, 0, intPtr(1), nil, uintPtr(5), intPtr(1))
	if err != nil {
		t.Errorf("expected no error, got: %v", err)
	}
}

func TestValidateCampaignType_FreeItem_MissingRewardProduct(t *testing.T) {
	err := validateCampaignType(entities.CampaignTypeBuyXProductGetYProductFree, 0, intPtr(1), nil, nil, intPtr(1))
	if err == nil {
		t.Error("expected error for missing reward_product_id")
	}
}

func TestValidateCampaignType_FreeItem_MissingRewardQty(t *testing.T) {
	err := validateCampaignType(entities.CampaignTypeBuyXProductGetYProductFree, 0, intPtr(1), nil, uintPtr(5), nil)
	if err == nil {
		t.Error("expected error for missing reward_quantity")
	}
}

func TestValidateCampaignType_UnknownType(t *testing.T) {
	err := validateCampaignType("something_invalid", 10, nil, nil, nil, nil)
	if err == nil {
		t.Error("expected error for unknown campaign type")
	}
}
