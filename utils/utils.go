package utils

import (
	"etop/db"
	"etop/models"
	"fmt"
	"net/http"
	"regexp"
	"strconv"
	"strings"
	"time"

	"crypto/sha1"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"

	"github.com/a-h/templ"
	"golang.org/x/crypto/bcrypt"
)

func ExtractUserMemberOfWorkspace(workspace models.Workspace) []models.UserRole {
	var members []models.UserRole
	for _, member := range workspace.Members {
		var m models.UserRole
		m.ID = member.User.ID
		m.FullName = member.User.FullName
		m.Role = member.Role
		m.Color = member.User.Color
		members = append(members, m)
	}
	return members
}

func ExtractUserMemberOfProject(project models.Project) []models.UserRole {
	var members []models.UserRole
	for _, member := range project.Members {
		var m models.UserRole
		m.ID = member.User.ID
		m.FullName = member.User.FullName
		m.Role = member.Role
		m.Color = member.User.Color
		members = append(members, m)
	}
	return members
}

func ExtractTagsFromProject(project models.Project) []string {
	var tags []string
	for _, projectTags := range project.Tags {
		tags = append(tags, projectTags.Tag)
	}
	return tags
}

// HoursDiff menghitung selisih jam (dibulatkan ke bawah) antara dua datetime
// menggunakan location tertentu (misal Asia/Jakarta).
func TimeDiff(start time.Time, to time.Time) time.Duration {
	loc, _ := time.LoadLocation("Asia/Jakarta")

	layout := "2006-01-02 15:04:05"

	s, _ := time.ParseInLocation(layout, start.Format(layout), loc)
	t, _ := time.ParseInLocation(layout, to.Format(layout), loc)

	// Konversi keduanya ke lokasi yg sama
	s2 := s.In(loc)
	t2 := t.In(loc)

	// Hitung selisih
	diff := t2.Sub(s2)
	return diff
}

// TimeAgo menerima t (waktu log) dan mengembalikan string relatif dibanding sekarang
func TimeAgo(tLog time.Time) string {
	loc, _ := time.LoadLocation("Asia/Jakarta")

	layout := "2006-01-02 15:04:05"
	now := time.Now().Local()
	s, _ := time.ParseInLocation(layout, tLog.Format(layout), loc)
	t, _ := time.ParseInLocation(layout, now.Format(layout), loc)

	// Konversi keduanya ke lokasi yg sama
	s2 := s.In(loc)
	t2 := t.In(loc)

	// Hitung selisih
	diff := t2.Sub(s2)

	// now := time.Now().Local()
	// diff := now.Sub(t.Local())

	seconds := int(diff.Seconds())
	minutes := int(diff.Minutes())
	hours := int(diff.Hours())
	days := int(diff.Hours() / 24)
	weeks := int(diff.Hours() / (24 * 7))
	months := int(diff.Hours() / (24 * 30))
	years := int(diff.Hours() / (24 * 365))

	switch {
	case seconds < 60:
		return "less than a minute ago"
	case minutes < 60:
		if minutes == 1 {
			return "1 minute ago"
		}
		return fmt.Sprintf("%d minutes ago", minutes)
	case hours < 24:
		if hours == 1 {
			return "1 hour ago"
		}
		return fmt.Sprintf("%d hours ago", hours)
	case days < 7:
		if days == 1 {
			return "yesterday"
		}
		return fmt.Sprintf("%d days ago", days)
	case weeks < 5:
		if weeks == 1 {
			return "1 week ago"
		}
		return fmt.Sprintf("%d weeks ago", weeks)
	case months < 12:
		if months == 1 {
			return "1 month ago"
		}
		return fmt.Sprintf("%d months ago", months)
	default:
		if years == 1 {
			return "1 year ago"
		}
		return fmt.Sprintf("%d years ago", years)
	}
}

func MustJSON(v any) []byte {
	b, _ := json.Marshal(v)
	return b
}

func FormatDate(t time.Time) string {
	if t.IsZero() {
		return ""
	}
	return t.Format("2006-01-02")
}

func FormatDateTime(t time.Time) string {
	if t.IsZero() {
		return ""
	}
	return t.Format("2006-01-02 15:04:05")
}

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

func DbMutation() {

	projectStatuses := []models.ProjectStatus{}
	err := db.PgSql.Find(&projectStatuses).Error

	if len(projectStatuses) == 0 || err != nil {
		projectStatuses = append(projectStatuses, models.ProjectStatus{Status: "PLANNING", Label: "Planning", Color: "bg-blue-300 text-blue-800 dark:bg-blue-400/30 dark:text-blue-100", Form: 1, Value: 3})
		projectStatuses = append(projectStatuses, models.ProjectStatus{Status: "IN_PROGRESS", Label: "In Progress", Color: "bg-purple-300 text-purple-800 dark:bg-purple-400/30 dark:text-purple-100", Form: 1, Value: 4})
		projectStatuses = append(projectStatuses, models.ProjectStatus{Status: "ON_HOLD", Label: "On Hold", Color: "bg-yellow-300 text-yellow-800 dark:bg-yellow-400/30 dark:text-yellow-100", Form: 0, Value: 2})
		projectStatuses = append(projectStatuses, models.ProjectStatus{Status: "COMPLETED", Label: "Completed", Color: "bg-green-300 text-green-800 dark:bg-green-400/30 dark:text-green-100", Form: 1, Value: 5})
		projectStatuses = append(projectStatuses, models.ProjectStatus{Status: "CANCELLED", Label: "Cancelled", Color: "bg-red-300 text-red-800 dark:bg-red-400/30 dark:text-red-100", Form: 0, Value: 0})
		if err := db.PgSql.Create(&projectStatuses).Error; err != nil {
			println(err.Error())
		}
	}

	taskStatuses := []models.TaskStatus{}
	err = db.PgSql.Find(&taskStatuses).Error

	if len(taskStatuses) == 0 || err != nil {
		taskStatuses = append(taskStatuses, models.TaskStatus{Status: "TO_DO", Label: "To Do", Color: "bg-red-400 text-primary dark:bg-red-500", Form: 1, Value: 2, Level: 1})
		taskStatuses = append(taskStatuses, models.TaskStatus{Status: "IN_PROGRESS", Label: "In Progress", Color: "bg-yellow-400 text-primary dark:bg-yellow-500", Form: 1, Value: 3, Level: 2})
		taskStatuses = append(taskStatuses, models.TaskStatus{Status: "IN_REVIEW", Label: "In Review", Color: "bg-blue-400 text-primary dark:bg-blue-500", Form: 0, Value: 4, Level: 2})
		taskStatuses = append(taskStatuses, models.TaskStatus{Status: "DONE", Label: "Done", Color: "bg-emerald-400 text-primary dark:bg-emerald-500", Form: 1, Value: 5, Level: 3})
		taskStatuses = append(taskStatuses, models.TaskStatus{Status: "CANCELLED", Label: "Cancelled", Color: "bg-pink-400 text-primary dark:bg-pink-500", Form: 0, Value: 0, Level: 3})
		if err := db.PgSql.Create(&taskStatuses).Error; err != nil {
			println(err.Error())
		}
	}

	taskPriorities := []models.TaskPriority{}
	err = db.PgSql.Find(&taskPriorities).Error

	if len(taskPriorities) == 0 || err != nil {
		taskPriorities = append(taskPriorities, models.TaskPriority{Priority: "LOW", Label: "Low", Color: "bg-green-400 dark:bg-green-600/30", Value: 1, Level: 1})
		taskPriorities = append(taskPriorities, models.TaskPriority{Priority: "MEDIUM", Label: "Medium", Color: "bg-yellow-400 dark:bg-yellow-600/30 ", Value: 3, Level: 2})
		taskPriorities = append(taskPriorities, models.TaskPriority{Priority: "HIGH", Label: "High", Color: "bg-red-400 dark:bg-red-600/30", Value: 5, Level: 3})
		if err := db.PgSql.Create(&taskPriorities).Error; err != nil {
			println(err.Error())
		}
	}

	settings := []models.Setting{}
	err = db.PgSql.Find(&settings).Error

	if len(settings) == 0 || err != nil {
		settings = append(settings, models.Setting{ID: GenerateHash(strconv.Itoa(1)), Name: "updated_task", Label: "Task Updates"})
		settings = append(settings, models.Setting{ID: GenerateHash(strconv.Itoa(2)), Name: "updated_project", Label: "Project Updates"})
		if err := db.PgSql.Create(&settings).Error; err != nil {
			println(err.Error())
		}
	}
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
		"A": "bg-red-500 text-white",
		"B": "bg-blue-500 text-white",
		"C": "bg-green-500 text-white",
		"D": "bg-yellow-500 text-black",
		"E": "bg-purple-500 text-white",
		"F": "bg-pink-500 text-white",
		"G": "bg-indigo-500 text-white",
		"H": "bg-teal-500 text-white",
		"I": "bg-cyan-500 text-white",
		"J": "bg-lime-500 text-black",
		"K": "bg-amber-500 text-black",
		"L": "bg-orange-500 text-white",
		"M": "bg-emerald-500 text-white",
		"N": "bg-fuchsia-500 text-white",
		"O": "bg-rose-500 text-white",
		"P": "bg-sky-500 text-white",
		"Q": "bg-violet-500 text-white",
		"R": "bg-stone-500 text-white",
		"S": "bg-zinc-500 text-white",
		"T": "bg-neutral-500 text-white",
		"U": "bg-slate-500 text-white",
		"V": "bg-gray-500 text-white",
		"W": "bg-blue-400 text-white",
		"X": "bg-red-400 text-white",
		"Y": "bg-green-400 text-white",
		"Z": "bg-yellow-400 text-black",
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
			class := "bg-" + c + "-" + fmt.Sprint(s) + " " + text
			buckets = append(buckets, class)
		}
	}
	return buckets
}

// TailwindForUsername memilih warna sesuai username
func TailwindForUsername(username string) string {
	if strings.Trim(username, " ") == "" {
		return "bg-gray-400 text-black dark:bg-gray-500 dark:text-white"
	}
	var tailwindBuckets = GenerateBuckets()

	hash := sha1.Sum([]byte(strings.Trim(username, " ")))
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
