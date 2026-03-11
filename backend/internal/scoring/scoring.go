package scoring

import "github.com/rhousand/svcompare/internal/models"

// Section defines a scoring category with its weight and associated question IDs.
type Section struct {
	Name        string
	Weight      float64
	QuestionIDs []int
}

// Sections is the canonical list of scoring sections matching the sailboat buyer's guide.
// This is the single source of truth for backend scoring logic.
var Sections = []Section{
	{Name: "Ownership & History", Weight: 0.20, QuestionIDs: []int{1, 2, 3, 4}},
	{Name: "Engine & Mechanical", Weight: 0.20, QuestionIDs: []int{5, 6, 7}},           // Q8 removed
	{Name: "Sails & Rig", Weight: 0.15, QuestionIDs: []int{9, 10, 11, 15, 26}},         // Q15 moved here; Q26 = running rigging
	{Name: "Systems", Weight: 0.15, QuestionIDs: []int{12, 13, 14}},                    // Q15 moved to Sails & Rig
	{Name: "Survey & Hull Condition", Weight: 0.20, QuestionIDs: []int{16, 17, 18, 19}},
	{Name: "Electronics & Safety", Weight: 0.10, QuestionIDs: []int{20, 21, 22}},
	{Name: "Transaction", Weight: 0.00, QuestionIDs: []int{23, 24, 25}}, // informational only
}

// Calculate computes weighted scores for a boat given its scores.
// Unscored questions (Value == nil) are excluded from section averages.
// The Transaction section (weight 0) is included in results but not in TotalWeighted.
func Calculate(boat models.Boat) models.BoatResult {
	// Build score lookup: questionID -> value
	scoreMap := make(map[int]*int, len(boat.Scores))
	for _, s := range boat.Scores {
		scoreMap[s.QuestionID] = s.Value
	}

	sections := make([]models.SectionResult, 0, len(Sections))
	var totalWeighted float64

	for _, sec := range Sections {
		var sum float64
		var count int
		for _, qid := range sec.QuestionIDs {
			if v, ok := scoreMap[qid]; ok && v != nil {
				sum += float64(*v)
				count++
			}
		}

		var rawAvg float64
		if count > 0 {
			rawAvg = sum / float64(count)
		}
		weighted := rawAvg * sec.Weight

		sections = append(sections, models.SectionResult{
			Name:          sec.Name,
			Weight:        sec.Weight,
			RawAverage:    rawAvg,
			WeightedScore: weighted,
			ScoredCount:   count,
			TotalCount:    len(sec.QuestionIDs),
		})

		if sec.Weight > 0 {
			totalWeighted += weighted
		}
	}

	return models.BoatResult{
		Boat:          boat,
		Sections:      sections,
		TotalWeighted: totalWeighted,
	}
}
