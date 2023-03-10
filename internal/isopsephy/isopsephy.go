// Package isopsephy calculates the "number" of Greek words or phrases.
package isopsephy

import "fmt"

// Calculate finds the "number" of a word by summing the numeric equivalent of
// each Greek letter. If an unrecognized character is encountered an error is
// returned.
func Calculate(word string) (int, error) {
	total := 0
	for _, r := range word {
		switch r {
		case ' ', '῎', '῞', '῾', '῏', '᾿', '῟', 0x301, 0x302, 0x304, 0x313, 0x314, 0x342:
		case 'Α', 'Ἀ', 'α', 'ά', 'ᾶ', 'ἁ':
			total += 1
		case 'Β', 'β':
			total += 2
		case 'Γ', 'γ':
			total += 3
		case 'Δ', 'δ':
			total += 4
		case 'Ε', 'Ἐ', 'ε', 'έ', 'ἑ':
			total += 5
		case 'Ϝ', 'ϝ', 'Ϛ', 'ϛ', 'ς':
			total += 6
		case 'Ζ', 'ζ':
			total += 7
		case 'Η', 'Ἠ', 'η', 'ή', 'ῆ', 'ἡ':
			total += 8
		case 'Θ', 'θ':
			total += 9
		case 'Ι', 'Ἰ', 'ι', 'ί', 'ῖ', 'ἰ', 'ΐ', 'ϊ', 'ἴ', 'ἱ', 'ἵ':
			total += 10
		case 'Κ', 'κ':
			total += 20
		case 'Λ', 'λ':
			total += 30
		case 'Μ', 'μ':
			total += 40
		case 'Ν', 'Ͷ', 'ͷ', 'ν':
			total += 50
		case 'Ξ', 'ξ':
			total += 60
		case 'Ο', 'Ὀ', 'ο', 'ό', 'ὀ', 'ὁ', 'ὄ':
			total += 70
		case 'Π', 'π':
			total += 80
		case 'Ϙ', 'ϙ':
			total += 90
		case 'Ρ', 'Ῥ', 'ρ':
			total += 100
		case 'Σ', 'σ':
			total += 200
		case 'Τ', 'τ':
			total += 300
		case 'Υ', 'υ', 'ύ', 'ὐ', 'ῦ', 'ὔ', 'ϋ', 'ὑ', 'ΰ', 'ὕ':
			total += 400
		case 'Φ', 'φ':
			total += 500
		case 'Χ', 'χ':
			total += 600
		case 'Ψ', 'ψ':
			total += 700
		case 'Ω', 'Ὠ', 'ω', 'ώ', 'ῶ', 'ὡ':
			total += 800
		case 'Ϡ', 'ϡ':
			total += 900
		default:
			return 0, fmt.Errorf("bad character %s (%x)", string(r), r)
		}
	}

	return total, nil
}
