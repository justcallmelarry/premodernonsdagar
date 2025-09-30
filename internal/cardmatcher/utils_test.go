package cardmatcher

import (
	"testing"
)

func TestNormalizeString(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "simple lowercase",
			input:    "Lightning Bolt",
			expected: "lightning bolt",
		},
		{
			name:     "already lowercase",
			input:    "lightning bolt",
			expected: "lightning bolt",
		},
		{
			name:     "with punctuation",
			input:    "Swords to Plowshares",
			expected: "swords to plowshares",
		},
		{
			name:     "with special characters",
			input:    "Force of Will!",
			expected: "force of will",
		},
		{
			name:     "with numbers",
			input:    "Lightning Bolt 3",
			expected: "lightning bolt 3",
		},
		{
			name:     "with hyphens and apostrophes",
			input:    "Jace, the Mind-Sculptor's Ability",
			expected: "jace the mindsculptors ability",
		},
		{
			name:     "multiple spaces",
			input:    "Lightning    Bolt",
			expected: "lightning bolt",
		},
		{
			name:     "leading/trailing whitespace",
			input:    "  Lightning Bolt  ",
			expected: "lightning bolt",
		},
		{
			name:     "empty string",
			input:    "",
			expected: "",
		},
		{
			name:     "only whitespace",
			input:    "   ",
			expected: "",
		},
		{
			name:     "only punctuation",
			input:    "!@#$%",
			expected: "",
		},
		{
			name:     "mixed unicode characters",
			input:    "Ætherling's Power",
			expected: "ætherlings power",
		},
		{
			name:     "numbers and letters",
			input:    "X Spell 123",
			expected: "x spell 123",
		},
		{
			name:     "with leading count",
			input:    "3 Wrath of God",
			expected: "wrath of god",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := normalizeString(tt.input)
			if result != tt.expected {
				t.Errorf("normalizeString(%q) = %q, want %q", tt.input, result, tt.expected)
			}
		})
	}
}

func TestLevenshteinDistance(t *testing.T) {
	tests := []struct {
		name     string
		a        string
		b        string
		expected int
	}{
		{
			name:     "identical strings",
			a:        "hello",
			b:        "hello",
			expected: 0,
		},
		{
			name:     "empty strings",
			a:        "",
			b:        "",
			expected: 0,
		},
		{
			name:     "one empty string",
			a:        "hello",
			b:        "",
			expected: 5,
		},
		{
			name:     "other empty string",
			a:        "",
			b:        "world",
			expected: 5,
		},
		{
			name:     "single character substitution",
			a:        "cat",
			b:        "bat",
			expected: 1,
		},
		{
			name:     "single character insertion",
			a:        "cat",
			b:        "cart",
			expected: 1,
		},
		{
			name:     "single character deletion",
			a:        "cart",
			b:        "cat",
			expected: 1,
		},
		{
			name:     "multiple operations",
			a:        "kitten",
			b:        "sitting",
			expected: 3,
		},
		{
			name:     "completely different",
			a:        "abc",
			b:        "xyz",
			expected: 3,
		},
		{
			name:     "case sensitive",
			a:        "Hello",
			b:        "hello",
			expected: 1,
		},
		{
			name:     "longer strings",
			a:        "Lightning Bolt",
			b:        "Lightining Bolt",
			expected: 1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := levenshteinDistance(tt.a, tt.b)
			if result != tt.expected {
				t.Errorf("levenshteinDistance(%q, %q) = %d, want %d", tt.a, tt.b, result, tt.expected)
			}
		})
	}
}

func TestCalculateSimilarity(t *testing.T) {
	tests := []struct {
		name               string
		original           string
		target             string
		expectedSimilarity float64
		expectedDistance   int
		tolerance          float64 // for floating point comparison
	}{
		{
			name:               "exact match",
			original:           "Lightning Bolt",
			target:             "Lightning Bolt",
			expectedSimilarity: 1.0,
			expectedDistance:   0,
			tolerance:          0.001,
		},
		{
			name:               "case insensitive match",
			original:           "Lightning Bolt",
			target:             "lightning bolt",
			expectedSimilarity: 1.0,
			expectedDistance:   0,
			tolerance:          0.001,
		},
		{
			name:               "punctuation normalization",
			original:           "Swords to Plowshares",
			target:             "swords to plowshares",
			expectedSimilarity: 1.0,
			expectedDistance:   0,
			tolerance:          0.001,
		},
		{
			name:             "single typo",
			original:         "Lightning Bolt",
			target:           "Lightining Bolt",
			expectedDistance: 1,
			tolerance:        0.001,
		},
		{
			name:      "substring match gets bonus",
			original:  "Lightning",
			target:    "Lightning Bolt",
			tolerance: 0.001,
		},
		{
			name:      "reverse substring match gets bonus",
			original:  "Lightning Bolt",
			target:    "Lightning",
			tolerance: 0.001,
		},
		{
			name:      "completely different",
			original:  "Lightning Bolt",
			target:    "Black Lotus",
			tolerance: 0.001,
		},
		{
			name:               "empty strings",
			original:           "",
			target:             "",
			expectedSimilarity: 1.0,
			expectedDistance:   0,
			tolerance:          0.001,
		},
		{
			name:      "one empty string",
			original:  "Lightning Bolt",
			target:    "",
			tolerance: 0.001,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			similarity, distance := calculateSimilarity(tt.original, tt.target)

			// Check distance if expected
			if tt.expectedDistance > 0 {
				if distance != tt.expectedDistance {
					t.Errorf("calculateSimilarity(%q, %q) distance = %d, want %d",
						tt.original, tt.target, distance, tt.expectedDistance)
				}
			}

			// Check similarity if expected
			if tt.expectedSimilarity > 0 {
				if abs(similarity-tt.expectedSimilarity) > tt.tolerance {
					t.Errorf("calculateSimilarity(%q, %q) similarity = %f, want %f (±%f)",
						tt.original, tt.target, similarity, tt.expectedSimilarity, tt.tolerance)
				}
			}

			// Basic sanity checks
			if similarity < 0.0 || similarity > 1.0 {
				t.Errorf("calculateSimilarity(%q, %q) similarity = %f, should be between 0.0 and 1.0",
					tt.original, tt.target, similarity)
			}

			if distance < 0 {
				t.Errorf("calculateSimilarity(%q, %q) distance = %d, should be >= 0",
					tt.original, tt.target, distance)
			}
		})
	}
}

func TestMin(t *testing.T) {
	tests := []struct {
		name     string
		a, b, c  int
		expected int
	}{
		{
			name:     "first is minimum",
			a:        1,
			b:        2,
			c:        3,
			expected: 1,
		},
		{
			name:     "second is minimum",
			a:        2,
			b:        1,
			c:        3,
			expected: 1,
		},
		{
			name:     "third is minimum",
			a:        3,
			b:        2,
			c:        1,
			expected: 1,
		},
		{
			name:     "all equal",
			a:        5,
			b:        5,
			c:        5,
			expected: 5,
		},
		{
			name:     "negative numbers",
			a:        -1,
			b:        -2,
			c:        -3,
			expected: -3,
		},
		{
			name:     "zeros",
			a:        0,
			b:        1,
			c:        2,
			expected: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := min(tt.a, tt.b, tt.c)
			if result != tt.expected {
				t.Errorf("min(%d, %d, %d) = %d, want %d", tt.a, tt.b, tt.c, result, tt.expected)
			}
		})
	}
}

func TestMax(t *testing.T) {
	tests := []struct {
		name     string
		a, b     int
		expected int
	}{
		{
			name:     "first is maximum",
			a:        3,
			b:        2,
			expected: 3,
		},
		{
			name:     "second is maximum",
			a:        2,
			b:        3,
			expected: 3,
		},
		{
			name:     "equal values",
			a:        5,
			b:        5,
			expected: 5,
		},
		{
			name:     "negative numbers",
			a:        -1,
			b:        -2,
			expected: -1,
		},
		{
			name:     "zero and positive",
			a:        0,
			b:        1,
			expected: 1,
		},
		{
			name:     "zero and negative",
			a:        0,
			b:        -1,
			expected: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := max(tt.a, tt.b)
			if result != tt.expected {
				t.Errorf("max(%d, %d) = %d, want %d", tt.a, tt.b, result, tt.expected)
			}
		})
	}
}

// Helper function for floating point comparison
func abs(f float64) float64 {
	if f < 0 {
		return -f
	}
	return f
}

// Benchmark tests for performance
func BenchmarkNormalizeString(b *testing.B) {
	testString := "Lightning Bolt's Amazing Power!"
	for i := 0; i < b.N; i++ {
		normalizeString(testString)
	}
}

func BenchmarkLevenshteinDistance(b *testing.B) {
	a := "Lightning Bolt"
	c := "Lightining Bolt"
	for i := 0; i < b.N; i++ {
		levenshteinDistance(a, c)
	}
}

func BenchmarkCalculateSimilarity(b *testing.B) {
	original := "Lightning Bolt"
	target := "Lightining Bolt"
	for i := 0; i < b.N; i++ {
		calculateSimilarity(original, target)
	}
}
