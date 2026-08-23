package chatbot

import (
	"regexp"
	"sync"
	"time"
)

type Session struct {
	LastIntent string
	LastMenu   string
	LastUpdate time.Time
}

var sessionStore = make(map[string]*Session)
var sessionMutex sync.RWMutex

func getSession(id string) *Session {
	if id == "" {
		return &Session{}
	}
	sessionMutex.Lock()
	defer sessionMutex.Unlock()
	if sess, ok := sessionStore[id]; ok {
		sess.LastUpdate = time.Now()
		return sess
	}
	sess := &Session{LastUpdate: time.Now()}
	sessionStore[id] = sess
	return sess
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

func levenshtein(s1, s2 string) int {
	if len(s1) == 0 {
		return len(s2)
	}
	if len(s2) == 0 {
		return len(s1)
	}
	d := make([][]int, len(s1)+1)
	for i := range d {
		d[i] = make([]int, len(s2)+1)
		d[i][0] = i
	}
	for j := 0; j <= len(s2); j++ {
		d[0][j] = j
	}
	for i := 1; i <= len(s1); i++ {
		for j := 1; j <= len(s2); j++ {
			cost := 1
			if s1[i-1] == s2[j-1] {
				cost = 0
			}
			d[i][j] = min(min(d[i-1][j]+1, d[i][j-1]+1), d[i-1][j-1]+cost)
		}
	}
	return d[len(s1)][len(s2)]
}

// ExtractYear finds a 4-digit year in the text
func extractYear(text string) (int, bool) {
	re := regexp.MustCompile(`\b(20\d{2})\b`)
	match := re.FindString(text)
	if match != "" {
		var year int
		importFmt := false
		_ = importFmt // hack to allow fmt usage via a loop if needed, but we can just loop digits
		year = 0
		for _, c := range match {
			year = year*10 + int(c-'0')
		}
		return year, true
	}
	return 0, false
}

type IntentRule struct {
	Intent   string
	Keywords []string
}

var rules = []IntentRule{
	{"dalkot_spj", []string{"spj", "petugas spj", "laporan spj"}},
	{"dalkot_riil", []string{"riil", "rill", "petugas riil", "petugas rill"}},
	{"dalkot_status", []string{"proses", "selesai", "sukses", "status dalkot", "progres", "draft", "tertahan"}},
	{"dalkot_finance", []string{"dalkot", "dalam kota", "perjalanan dinas dalam kota"}},
	{"gup_menus", []string{"12 menu", "menu gup", "jenis gup", "pilihan gup"}},
	{"gup_overview", []string{"gup", "anggaran", "realisasi", "serapan", "keuangan", "pagu", "uang"}},
}
