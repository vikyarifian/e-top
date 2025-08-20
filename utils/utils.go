package utils

import (
	"net/http"
	"strings"

	"github.com/a-h/templ"
)

func Render(w http.ResponseWriter, r *http.Request, c templ.Component) error {
	return c.Render(r.Context(), w)
}

func LetterToColor(letter string) string {
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

	// Normalize to uppercase
	letter = string(strings.ToUpper(string(letter))[0])

	// Get color or return a default
	if color, ok := colorMap[letter]; ok {
		return color
	}
	return "black" // default if not A-Z
}
