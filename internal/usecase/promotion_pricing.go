package usecase

import (
	"fmt"
	"math"

	"github.com/usernamesalah/rh-pos/internal/domain/entities"
)

type pricingLineInput struct {
	ProductID  uint
	Quantity   int
	UnitPrice  float64
	IsFreeItem bool
	CampaignID *uint
}

type pricingLineResult struct {
	ProductID          uint
	Quantity           int
	UnitPrice          float64
	DiscountPercentage float64
	DiscountAmount     float64
	CampaignID         *uint
	CampaignType       string
	IsFreeItem         bool
	CampaignGroupKey   string
}

type promoCandidate struct {
	campaign      entities.DiscountCampaign
	discountValue float64
	results       []pricingLineResult
}

func evaluatePromotions(purchasedLines []pricingLineInput, freeLines []pricingLineInput, campaigns []entities.DiscountCampaign) ([]pricingLineResult, error) {
	campaignProductMap := buildCampaignProductMap(campaigns)
	finalResults := make([]pricingLineResult, 0, len(purchasedLines)+len(freeLines))

	for _, line := range purchasedLines {
		eligibleCampaigns := findEligibleCampaigns(line.ProductID, campaigns, campaignProductMap)

		if len(eligibleCampaigns) == 0 {
			finalResults = append(finalResults, pricingLineResult{
				ProductID: line.ProductID,
				Quantity:  line.Quantity,
				UnitPrice: line.UnitPrice,
			})
			continue
		}

		var bestCandidate *promoCandidate
		for i := range eligibleCampaigns {
			c := eligibleCampaigns[i]
			candidate := evaluateSingleCampaign(c, line, freeLines)
			if candidate == nil {
				continue
			}
			if bestCandidate == nil || candidate.discountValue > bestCandidate.discountValue {
				bestCandidate = candidate
			} else if candidate.discountValue == bestCandidate.discountValue {
				bestCandidate = breakTie(bestCandidate, candidate)
			}
		}

		if bestCandidate == nil {
			finalResults = append(finalResults, pricingLineResult{
				ProductID: line.ProductID,
				Quantity:  line.Quantity,
				UnitPrice: line.UnitPrice,
			})
			continue
		}

		finalResults = append(finalResults, bestCandidate.results...)
	}

	if err := validateUnclaimedFreeLines(freeLines, finalResults); err != nil {
		return nil, err
	}

	return finalResults, nil
}

func buildCampaignProductMap(campaigns []entities.DiscountCampaign) map[uint]map[uint]bool {
	m := make(map[uint]map[uint]bool)
	for _, c := range campaigns {
		pids := make(map[uint]bool)
		for _, p := range c.Products {
			pids[p.ProductID] = true
		}
		m[c.ID] = pids
	}
	return m
}

func findEligibleCampaigns(productID uint, campaigns []entities.DiscountCampaign, campaignProductMap map[uint]map[uint]bool) []entities.DiscountCampaign {
	var result []entities.DiscountCampaign
	for _, c := range campaigns {
		if pids, ok := campaignProductMap[c.ID]; ok && pids[productID] {
			result = append(result, c)
		}
	}
	return result
}

func evaluateSingleCampaign(campaign entities.DiscountCampaign, line pricingLineInput, freeLines []pricingLineInput) *promoCandidate {
	switch campaign.CampaignType {
	case entities.CampaignTypeProductPercentageDiscount:
		return evalProductPercentageDiscount(campaign, line)
	case entities.CampaignTypeBuyXQtyGetDiscountAmount:
		return evalBuyXQtyGetDiscountAmount(campaign, line)
	case entities.CampaignTypeBuyXQtyGetDiscountPercent:
		return evalBuyXQtyGetDiscountPercent(campaign, line)
	case entities.CampaignTypeBuyXProductGetYProductFree:
		return evalBuyXProductGetYFree(campaign, line, freeLines)
	default:
		return nil
	}
}

func evalProductPercentageDiscount(campaign entities.DiscountCampaign, line pricingLineInput) *promoCandidate {
	pct := campaign.DiscountPercentage
	totalDiscount := line.UnitPrice * float64(line.Quantity) * pct / 100
	discountedUnit := roundTo2(line.UnitPrice * (1 - pct/100))

	return &promoCandidate{
		campaign:      campaign,
		discountValue: totalDiscount,
		results: []pricingLineResult{{
			ProductID:          line.ProductID,
			Quantity:           line.Quantity,
			UnitPrice:          discountedUnit,
			DiscountPercentage: pct,
			DiscountAmount:     roundTo2(totalDiscount),
			CampaignID:         &campaign.ID,
			CampaignType:       campaign.CampaignType,
			CampaignGroupKey:   fmt.Sprintf("c%d-p%d", campaign.ID, line.ProductID),
		}},
	}
}

func evalBuyXQtyGetDiscountAmount(campaign entities.DiscountCampaign, line pricingLineInput) *promoCandidate {
	if campaign.BuyQuantity == nil || campaign.DiscountAmount == nil {
		return nil
	}
	buyQty := *campaign.BuyQuantity
	discAmt := *campaign.DiscountAmount
	if line.Quantity < buyQty {
		return nil
	}

	applications := line.Quantity / buyQty
	totalDiscount := float64(applications) * discAmt
	originalTotal := line.UnitPrice * float64(line.Quantity)

	if totalDiscount > originalTotal {
		totalDiscount = originalTotal
	}

	finalTotal := originalTotal - totalDiscount
	finalUnit := roundTo2(finalTotal / float64(line.Quantity))

	return &promoCandidate{
		campaign:      campaign,
		discountValue: totalDiscount,
		results: []pricingLineResult{{
			ProductID:        line.ProductID,
			Quantity:         line.Quantity,
			UnitPrice:        finalUnit,
			DiscountAmount:   roundTo2(totalDiscount),
			CampaignID:       &campaign.ID,
			CampaignType:     campaign.CampaignType,
			CampaignGroupKey: fmt.Sprintf("c%d-p%d", campaign.ID, line.ProductID),
		}},
	}
}

func evalBuyXQtyGetDiscountPercent(campaign entities.DiscountCampaign, line pricingLineInput) *promoCandidate {
	if campaign.BuyQuantity == nil {
		return nil
	}
	buyQty := *campaign.BuyQuantity
	pct := campaign.DiscountPercentage
	if line.Quantity < buyQty {
		return nil
	}

	applications := line.Quantity / buyQty
	eligibleUnits := applications * buyQty
	remainingUnits := line.Quantity - eligibleUnits

	discountOnEligible := float64(eligibleUnits) * line.UnitPrice * pct / 100
	eligibleTotal := float64(eligibleUnits)*line.UnitPrice - discountOnEligible
	remainingTotal := float64(remainingUnits) * line.UnitPrice
	finalTotal := eligibleTotal + remainingTotal
	finalUnit := roundTo2(finalTotal / float64(line.Quantity))

	return &promoCandidate{
		campaign:      campaign,
		discountValue: discountOnEligible,
		results: []pricingLineResult{{
			ProductID:          line.ProductID,
			Quantity:           line.Quantity,
			UnitPrice:          finalUnit,
			DiscountPercentage: pct,
			DiscountAmount:     roundTo2(discountOnEligible),
			CampaignID:         &campaign.ID,
			CampaignType:       campaign.CampaignType,
			CampaignGroupKey:   fmt.Sprintf("c%d-p%d", campaign.ID, line.ProductID),
		}},
	}
}

func evalBuyXProductGetYFree(campaign entities.DiscountCampaign, line pricingLineInput, freeLines []pricingLineInput) *promoCandidate {
	if campaign.BuyQuantity == nil || campaign.RewardProductID == nil || campaign.RewardQuantity == nil {
		return nil
	}
	buyQty := *campaign.BuyQuantity
	rewardPID := *campaign.RewardProductID
	rewardQty := *campaign.RewardQuantity

	if line.Quantity < buyQty {
		return nil
	}

	applications := line.Quantity / buyQty
	allowedFreeQty := applications * rewardQty

	var matchedFreeLine *pricingLineInput
	for i := range freeLines {
		if freeLines[i].IsFreeItem && freeLines[i].CampaignID != nil && *freeLines[i].CampaignID == campaign.ID && freeLines[i].ProductID == rewardPID {
			matchedFreeLine = &freeLines[i]
			break
		}
	}

	if matchedFreeLine == nil {
		return nil
	}

	actualFreeQty := matchedFreeLine.Quantity
	if actualFreeQty > allowedFreeQty {
		return nil
	}

	freeItemValue := matchedFreeLine.UnitPrice * float64(actualFreeQty)
	groupKey := fmt.Sprintf("c%d-p%d-r%d", campaign.ID, line.ProductID, rewardPID)

	results := []pricingLineResult{
		{
			ProductID:        line.ProductID,
			Quantity:         line.Quantity,
			UnitPrice:        line.UnitPrice,
			CampaignID:       &campaign.ID,
			CampaignType:     campaign.CampaignType,
			CampaignGroupKey: groupKey,
		},
		{
			ProductID:        rewardPID,
			Quantity:         actualFreeQty,
			UnitPrice:        0,
			DiscountAmount:   roundTo2(freeItemValue),
			CampaignID:       &campaign.ID,
			CampaignType:     campaign.CampaignType,
			IsFreeItem:       true,
			CampaignGroupKey: groupKey,
		},
	}

	return &promoCandidate{
		campaign:      campaign,
		discountValue: freeItemValue,
		results:       results,
	}
}

func breakTie(a, b *promoCandidate) *promoCandidate {
	typePriority := map[string]int{
		entities.CampaignTypeBuyXProductGetYProductFree: 0,
		entities.CampaignTypeBuyXQtyGetDiscountAmount:   1,
		entities.CampaignTypeBuyXQtyGetDiscountPercent:  2,
		entities.CampaignTypeProductPercentageDiscount:  3,
	}
	pa := typePriority[a.campaign.CampaignType]
	pb := typePriority[b.campaign.CampaignType]
	if pa < pb {
		return a
	}
	if pb < pa {
		return b
	}
	if a.campaign.ID < b.campaign.ID {
		return a
	}
	return b
}

func validateUnclaimedFreeLines(freeLines []pricingLineInput, results []pricingLineResult) error {
	claimedFree := make(map[uint]bool)
	for _, r := range results {
		if r.IsFreeItem {
			claimedFree[r.ProductID] = true
		}
	}
	for _, fl := range freeLines {
		if fl.IsFreeItem && !claimedFree[fl.ProductID] {
			return fmt.Errorf("free item for product %d is not valid for any active campaign", fl.ProductID)
		}
	}
	return nil
}

func roundTo2(v float64) float64 {
	return math.Round(v*100) / 100
}
