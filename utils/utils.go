package utils

import (
	"crypto/sha1"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"net/http"
	"regexp"
	"strings"

	"github.com/a-h/templ"
	"golang.org/x/crypto/bcrypt"
)

func Render(w http.ResponseWriter, r *http.Request, c templ.Component) error {
	return c.Render(r.Context(), w)
}

var emailRe = regexp.MustCompile(`^[^\s@]+@[^\s@]+\.[^\s@]+$`)

func IsEmailValidRegex(s string) bool {
	return emailRe.MatchString(s)
}

func GenerateHash(input string) string {
	input = strings.TrimSpace(input)
	sum := sha256.Sum256([]byte(input))
	return hex.EncodeToString(sum[:]) // 64 karakter hex
}

func HashPassword(password string) (string, error) {
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), 14)
	return string(bytes), err
}

func LetterToColorHex(letter string) string {
	// Define a map of letters to CSS colors
	colorMap := map[string]string{
		"A": "#ef4444", // red-500
		"B": "#3b82f6", // blue-500
		"C": "#22c55e", // green-500
		"D": "#eab308", // yellow-500
		"E": "#a855f7", // purple-500
		"F": "#ec4899", // pink-500
		"G": "#6366f1", // indigo-500
		"H": "#14b8a6", // teal-500
		"I": "#06b6d4", // cyan-500
		"J": "#84cc16", // lime-500
		"K": "#f59e0b", // amber-500
		"L": "#f97316", // orange-500
		"M": "#10b981", // emerald-500
		"N": "#d946ef", // fuchsia-500
		"O": "#f43f5e", // rose-500
		"P": "#0ea5e9", // sky-500
		"Q": "#8b5cf6", // violet-500
		"R": "#78716c", // stone-500
		"S": "#71717a", // zinc-500
		"T": "#737373", // neutral-500
		"U": "#64748b", // slate-500
		"V": "#6b7280", // gray-500
		"W": "#60a5fa", // blue-400
		"X": "#f87171", // red-400
		"Y": "#4ade80", // green-400
		"Z": "#fde047", // yellow-400
	}

	letter = string(strings.ToUpper(string(letter))[0])

	// Get color or return a default
	if color, ok := colorMap[letter]; ok {
		return color
	}
	return "#3b82f6" // default if not A-Z
}

func LetterToColor(letter string) string {
	// Peta huruf ke Tailwind class (dengan dark mode)
	colorMap := map[string]string{
		"A": "bg-red-500 dark:bg-gray-500 text-white",
		"B": "bg-blue-500 dark:bg-gray-500 text-white",
		"C": "bg-green-500 dark:bg-gray-500 text-white",
		"D": "bg-yellow-500 dark:bg-gray-500 text-black",
		"E": "bg-purple-500 dark:bg-gray-500 text-white",
		"F": "bg-pink-500 dark:bg-gray-500 text-white",
		"G": "bg-indigo-500 dark:bg-gray-500 text-white",
		"H": "bg-teal-500 dark:bg-gray-500 text-white",
		"I": "bg-cyan-500 dark:bg-gray-500 text-white",
		"J": "bg-lime-500 dark:bg-gray-500 text-black",
		"K": "bg-amber-500 dark:bg-gray-500 text-black",
		"L": "bg-orange-500 dark:bg-gray-500 text-white",
		"M": "bg-emerald-500 dark:bg-gray-500 text-white",
		"N": "bg-fuchsia-500 dark:bg-gray-500 text-white",
		"O": "bg-rose-500 dark:bg-gray-500 text-white",
		"P": "bg-sky-500 dark:bg-gray-500 text-white",
		"Q": "bg-violet-500 dark:bg-gray-500 text-white",
		"R": "bg-stone-500 dark:bg-gray-500 text-white",
		"S": "bg-zinc-500 dark:bg-gray-500 text-white",
		"T": "bg-neutral-500 dark:bg-gray-500 text-white",
		"U": "bg-slate-500 dark:bg-gray-500 text-white",
		"V": "bg-gray-500 dark:bg-gray-500 text-white",
		"W": "bg-blue-400 dark:bg-gray-500 text-white",
		"X": "bg-red-400 dark:bg-gray-500 text-white",
		"Y": "bg-green-400 dark:bg-gray-500 text-white",
		"Z": "bg-yellow-400 dark:bg-gray-500 text-black",
	}

	// Normalisasi ke uppercase dan ambil huruf pertama
	letter = strings.ToUpper(string(letter[0]))

	// Kembalikan class Tailwind
	if color, ok := colorMap[letter]; ok {
		return color
	}
	return "bg-blue-500 dark:bg-gray-500 text-white" // default
}

var baseColors = []string{
	"red", "orange", "amber", "yellow", "lime", "green", "emerald", "teal", "cyan", "sky",
	"blue", "indigo", "violet", "purple", "fuchsia", "pink", "rose",
	"slate", "gray", "zinc", "neutral", "stone",
}

var shades = []int{300, 400, 500, 600, 700}

func GenerateBuckets() []string {
	var buckets []string
	for _, c := range baseColors {
		for _, s := range shades {
			text := "text-black"
			if s >= 500 {
				text = "text-white"
			}
			class := "bg-" + c + "-" + fmt.Sprint(s) + " " + text +
				" dark:bg-gray-500 dark:text-white"
			buckets = append(buckets, class)
		}
	}
	return buckets
}

// TailwindForUsername memilih warna sesuai username
func TailwindForUsername(username string) string {
	if username == "" {
		return "bg-gray-400 text-black dark:bg-gray-500 dark:text-white"
	}
	var tailwindBuckets = GenerateBuckets()

	hash := sha1.Sum([]byte(username))
	idx := int(hash[0]) % len(tailwindBuckets)
	return tailwindBuckets[idx]
}

func ColorForLetter(letter string) string {
	if letter == "" {
		return "bg-gray-500 text-white dark:bg-gray-700"
	}

	hash := sha256.Sum256([]byte(letter))
	hexStr := hex.EncodeToString(hash[:])

	// ambil 6 karakter pertama → hex warna
	colorHex := hexStr[0:6]

	return fmt.Sprintf("#%s", colorHex)
}
