package usecase

import (
	"testing"

	"github.com/usernamesalah/rh-pos/internal/domain/entities"
)

func intPtr(v int) *int           { return &v }
func floatPtr(v float64) *float64 { return &v }
func uintPtr(v uint) *uint        { return &v }

func TestEvalProductPercentageDiscount(t *testing.T) {
	line := pricingLineInput{ProductID: 1, Quantity: 3, UnitPrice: 15000}
	campaigns := []entities.DiscountCampaign{{
		ID:                 10,
		CampaignType:       entities.CampaignTypeProductPercentageDiscount,
		DiscountPercentage: 20,
		Products:           []entities.DiscountCampaignProduct{{CampaignID: 10, ProductID: 1}},
	}}

	results, err := evaluatePromotions([]pricingLineInput{line}, nil, campaigns)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(results) != 1 {
		t.Fatalf("expected 1 result, got %d", len(results))
	}
	r := results[0]
	if r.UnitPrice != 12000 {
		t.Errorf("expected unit price 12000, got %v", r.UnitPrice)
	}
	if r.DiscountPercentage != 20 {
		t.Errorf("expected discount_percentage 20, got %v", r.DiscountPercentage)
	}
	if r.DiscountAmount != 9000 {
		t.Errorf("expected discount_amount 9000, got %v", r.DiscountAmount)
	}
	if r.CampaignID == nil || *r.CampaignID != 10 {
		t.Errorf("expected campaign_id 10")
	}
}

func TestBuyXQtyGetDiscountAmount_AppliesOnce(t *testing.T) {
	line := pricingLineInput{ProductID: 1, Quantity: 3, UnitPrice: 15000}
	campaigns := []entities.DiscountCampaign{{
		ID:             20,
		CampaignType:   entities.CampaignTypeBuyXQtyGetDiscountAmount,
		BuyQuantity:    intPtr(3),
		DiscountAmount: floatPtr(10000),
		Products:       []entities.DiscountCampaignProduct{{CampaignID: 20, ProductID: 1}},
	}}

	results, err := evaluatePromotions([]pricingLineInput{line}, nil, campaigns)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	r := results[0]
	// discount_amount is the authoritative value; unit price may have rounding drift
	if r.DiscountAmount != 10000 {
		t.Errorf("expected discount_amount 10000, got %v", r.DiscountAmount)
	}
	if r.CampaignID == nil || *r.CampaignID != 20 {
		t.Errorf("expected campaign_id 20")
	}
}

func TestBuyXQtyGetDiscountAmount_AppliesTwice(t *testing.T) {
	line := pricingLineInput{ProductID: 1, Quantity: 6, UnitPrice: 15000}
	campaigns := []entities.DiscountCampaign{{
		ID:             20,
		CampaignType:   entities.CampaignTypeBuyXQtyGetDiscountAmount,
		BuyQuantity:    intPtr(3),
		DiscountAmount: floatPtr(10000),
		Products:       []entities.DiscountCampaignProduct{{CampaignID: 20, ProductID: 1}},
	}}

	results, err := evaluatePromotions([]pricingLineInput{line}, nil, campaigns)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	r := results[0]
	if r.DiscountAmount != 20000 {
		t.Errorf("expected discount_amount 20000, got %v", r.DiscountAmount)
	}
}

func TestBuyXQtyGetDiscountAmount_InsufficientQty(t *testing.T) {
	line := pricingLineInput{ProductID: 1, Quantity: 2, UnitPrice: 15000}
	campaigns := []entities.DiscountCampaign{{
		ID:             20,
		CampaignType:   entities.CampaignTypeBuyXQtyGetDiscountAmount,
		BuyQuantity:    intPtr(3),
		DiscountAmount: floatPtr(10000),
		Products:       []entities.DiscountCampaignProduct{{CampaignID: 20, ProductID: 1}},
	}}

	results, err := evaluatePromotions([]pricingLineInput{line}, nil, campaigns)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	r := results[0]
	if r.CampaignID != nil {
		t.Error("expected no campaign applied for insufficient qty")
	}
	if r.UnitPrice != 15000 {
		t.Errorf("expected original price 15000, got %v", r.UnitPrice)
	}
}

func TestBuyXQtyGetDiscountPercent_PartialGroup(t *testing.T) {
	line := pricingLineInput{ProductID: 1, Quantity: 5, UnitPrice: 10000}
	campaigns := []entities.DiscountCampaign{{
		ID:                 30,
		CampaignType:       entities.CampaignTypeBuyXQtyGetDiscountPercent,
		BuyQuantity:        intPtr(3),
		DiscountPercentage: 10,
		Products:           []entities.DiscountCampaignProduct{{CampaignID: 30, ProductID: 1}},
	}}

	results, err := evaluatePromotions([]pricingLineInput{line}, nil, campaigns)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	r := results[0]
	// 3 eligible units: 3*10000*10% = 3000 discount
	// 2 remaining at full price
	// total = (30000-3000) + 20000 = 47000
	expectedDiscount := 3000.0
	if r.DiscountAmount != expectedDiscount {
		t.Errorf("expected discount_amount %v, got %v", expectedDiscount, r.DiscountAmount)
	}
	expectedTotal := 47000.0
	actualTotal := r.UnitPrice * float64(r.Quantity)
	if roundTo2(actualTotal) != expectedTotal {
		t.Errorf("expected total %v, got %v", expectedTotal, roundTo2(actualTotal))
	}
}

func TestBuyXGetYFree_ValidFreeLine(t *testing.T) {
	campaignID := uint(40)
	purchasedLine := pricingLineInput{ProductID: 1, Quantity: 2, UnitPrice: 50000}
	freeLine := pricingLineInput{ProductID: 2, Quantity: 2, UnitPrice: 20000, IsFreeItem: true, CampaignID: &campaignID}

	campaigns := []entities.DiscountCampaign{{
		ID:              campaignID,
		CampaignType:    entities.CampaignTypeBuyXProductGetYProductFree,
		BuyQuantity:     intPtr(1),
		RewardProductID: uintPtr(2),
		RewardQuantity:  intPtr(1),
		Products:        []entities.DiscountCampaignProduct{{CampaignID: campaignID, ProductID: 1}},
	}}

	results, err := evaluatePromotions([]pricingLineInput{purchasedLine}, []pricingLineInput{freeLine}, campaigns)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(results) != 2 {
		t.Fatalf("expected 2 result lines, got %d", len(results))
	}

	triggerLine := results[0]
	rewardLine := results[1]

	if triggerLine.ProductID != 1 || triggerLine.UnitPrice != 50000 {
		t.Errorf("trigger line unexpected: %+v", triggerLine)
	}
	if !rewardLine.IsFreeItem || rewardLine.UnitPrice != 0 || rewardLine.ProductID != 2 {
		t.Errorf("reward line unexpected: %+v", rewardLine)
	}
	if rewardLine.DiscountAmount != 40000 {
		t.Errorf("expected free item discount_amount 40000, got %v", rewardLine.DiscountAmount)
	}
}

func TestBuyXGetYFree_WrongProduct_Rejects(t *testing.T) {
	campaignID := uint(40)
	purchasedLine := pricingLineInput{ProductID: 1, Quantity: 1, UnitPrice: 50000}
	freeLine := pricingLineInput{ProductID: 99, Quantity: 1, UnitPrice: 10000, IsFreeItem: true, CampaignID: &campaignID}

	campaigns := []entities.DiscountCampaign{{
		ID:              campaignID,
		CampaignType:    entities.CampaignTypeBuyXProductGetYProductFree,
		BuyQuantity:     intPtr(1),
		RewardProductID: uintPtr(2),
		RewardQuantity:  intPtr(1),
		Products:        []entities.DiscountCampaignProduct{{CampaignID: campaignID, ProductID: 1}},
	}}

	_, err := evaluatePromotions([]pricingLineInput{purchasedLine}, []pricingLineInput{freeLine}, campaigns)
	if err == nil {
		t.Fatal("expected error for wrong free product, got nil")
	}
}

func TestBuyXGetYFree_ExcessiveFreeQty_Rejects(t *testing.T) {
	campaignID := uint(40)
	purchasedLine := pricingLineInput{ProductID: 1, Quantity: 1, UnitPrice: 50000}
	freeLine := pricingLineInput{ProductID: 2, Quantity: 5, UnitPrice: 20000, IsFreeItem: true, CampaignID: &campaignID}

	campaigns := []entities.DiscountCampaign{{
		ID:              campaignID,
		CampaignType:    entities.CampaignTypeBuyXProductGetYProductFree,
		BuyQuantity:     intPtr(1),
		RewardProductID: uintPtr(2),
		RewardQuantity:  intPtr(1),
		Products:        []entities.DiscountCampaignProduct{{CampaignID: campaignID, ProductID: 1}},
	}}

	_, err := evaluatePromotions([]pricingLineInput{purchasedLine}, []pricingLineInput{freeLine}, campaigns)
	if err == nil {
		t.Fatal("expected error for excessive free qty, got nil")
	}
}

func TestConflictResolution_ChoosesBestDiscount(t *testing.T) {
	line := pricingLineInput{ProductID: 1, Quantity: 3, UnitPrice: 15000}
	campaigns := []entities.DiscountCampaign{
		{
			ID:                 10,
			CampaignType:       entities.CampaignTypeProductPercentageDiscount,
			DiscountPercentage: 20,
			Products:           []entities.DiscountCampaignProduct{{CampaignID: 10, ProductID: 1}},
		},
		{
			ID:             20,
			CampaignType:   entities.CampaignTypeBuyXQtyGetDiscountAmount,
			BuyQuantity:    intPtr(3),
			DiscountAmount: floatPtr(10000),
			Products:       []entities.DiscountCampaignProduct{{CampaignID: 20, ProductID: 1}},
		},
	}

	results, err := evaluatePromotions([]pricingLineInput{line}, nil, campaigns)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	r := results[0]
	// percentage: 3*15000*20% = 9000
	// amount: 10000
	// buy_x_qty_get_discount_amount wins
	if r.CampaignID == nil || *r.CampaignID != 20 {
		t.Errorf("expected campaign 20 (amount) to win, got campaign %v", r.CampaignID)
	}
	if r.DiscountAmount != 10000 {
		t.Errorf("expected discount_amount 10000, got %v", r.DiscountAmount)
	}
}

func TestNoActiveCampaigns_RegularPricing(t *testing.T) {
	line := pricingLineInput{ProductID: 1, Quantity: 2, UnitPrice: 25000}

	results, err := evaluatePromotions([]pricingLineInput{line}, nil, nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	r := results[0]
	if r.CampaignID != nil {
		t.Error("expected no campaign")
	}
	if r.UnitPrice != 25000 {
		t.Errorf("expected original price 25000, got %v", r.UnitPrice)
	}
}

func TestFreeItemIgnored_WhenNoFreeLineSent(t *testing.T) {
	line := pricingLineInput{ProductID: 1, Quantity: 1, UnitPrice: 50000}
	campaigns := []entities.DiscountCampaign{
		{
			ID:                 10,
			CampaignType:       entities.CampaignTypeProductPercentageDiscount,
			DiscountPercentage: 10,
			Products:           []entities.DiscountCampaignProduct{{CampaignID: 10, ProductID: 1}},
		},
		{
			ID:              40,
			CampaignType:    entities.CampaignTypeBuyXProductGetYProductFree,
			BuyQuantity:     intPtr(1),
			RewardProductID: uintPtr(2),
			RewardQuantity:  intPtr(1),
			Products:        []entities.DiscountCampaignProduct{{CampaignID: 40, ProductID: 1}},
		},
	}

	results, err := evaluatePromotions([]pricingLineInput{line}, nil, campaigns)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	r := results[0]
	if r.CampaignID == nil || *r.CampaignID != 10 {
		t.Errorf("expected percentage campaign 10 to be used as fallback, got %v", r.CampaignID)
	}
}
