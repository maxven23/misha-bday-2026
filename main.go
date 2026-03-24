package main

import (
	"fmt"
	"math"
	"math/rand"
	"os"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// ─── colors ───────────────────────────────────────────────────────────────────

var (
	cOrange = lipgloss.Color("#FF8C00")
	cGold   = lipgloss.Color("#FFD700")
	cGreen  = lipgloss.Color("#00FF87")
	cBlue   = lipgloss.Color("#5AF")
	cPink   = lipgloss.Color("#FF69B4")
	cRed    = lipgloss.Color("#FF4444")
	cDim    = lipgloss.Color("#555")
	cWhite  = lipgloss.Color("#EFEFEF")

	sOrange = lipgloss.NewStyle().Foreground(cOrange)
	sGold   = lipgloss.NewStyle().Foreground(cGold).Bold(true)
	sGreen  = lipgloss.NewStyle().Foreground(cGreen)
	sBlue   = lipgloss.NewStyle().Foreground(cBlue)
	sRed    = lipgloss.NewStyle().Foreground(cRed)
	sDim    = lipgloss.NewStyle().Foreground(cDim)
	sWhite  = lipgloss.NewStyle().Foreground(cWhite)
	sBold   = lipgloss.NewStyle().Bold(true)
)

// ─── screens ──────────────────────────────────────────────────────────────────

type screen int

const (
	screenSplash screen = iota
	screenDeploy
	screenJokes
	screenGift
	screenGame
	screenWish
	screenFinal
)

// ─── ticks ────────────────────────────────────────────────────────────────────

type tickMsg struct{ t time.Time }
type deployStepMsg struct{}

func tickCmd() tea.Cmd {
	return tea.Tick(80*time.Millisecond, func(t time.Time) tea.Msg { return tickMsg{t} })
}
func deployTickCmd() tea.Cmd {
	return tea.Tick(120*time.Millisecond, func(t time.Time) tea.Msg { return deployStepMsg{} })
}

// ─── confetti ─────────────────────────────────────────────────────────────────

var confettiColors = []lipgloss.Color{cOrange, cGold, cGreen, cBlue, cPink, "#FF4500", "#9B59B6", "#E74C3C"}
var confettiChars = []rune{'*', '✦', '✧', '◆', '★', '✿', '❋', '·', '•', '♦', '♥'}

type particle struct {
	x, y int
	ch   rune
	col  lipgloss.Color
}

func spawnParticles(w, count int) []particle {
	pp := make([]particle, count)
	for i := range pp {
		pp[i] = particle{
			x:   rand.Intn(max(w, 1)),
			y:   0,
			ch:  confettiChars[rand.Intn(len(confettiChars))],
			col: confettiColors[rand.Intn(len(confettiColors))],
		}
	}
	return pp
}

func spawnParticlesScattered(w, h, count int) []particle {
	pp := make([]particle, count)
	for i := range pp {
		pp[i] = particle{
			x:   rand.Intn(max(w, 1)),
			y:   rand.Intn(max(h, 1)),
			ch:  confettiChars[rand.Intn(len(confettiChars))],
			col: confettiColors[rand.Intn(len(confettiColors))],
		}
	}
	return pp
}

func stepParticles(pp []particle, h int) []particle {
	out := pp[:0]
	for _, p := range pp {
		p.y++
		if p.y < h {
			out = append(out, p)
		}
	}
	return out
}

// ─── rockets ──────────────────────────────────────────────────────────────────

var sparkChars = []rune{'*', '+', '✦', '×', '◆', '·', '✧', '❋', '○'}

type spark struct {
	x, y   float64
	vx, vy float64
	col    lipgloss.Color
	ch     rune
	life   int
}

type rocket struct {
	x, y     float64
	vy       float64
	targetY  float64
	exploded bool
	sparks   []spark
}

func newRocket(w, h int) rocket {
	cx := float64(w/4 + rand.Intn(w/2))
	return rocket{
		x:       cx,
		y:       float64(h - 1),
		vy:      -(2.5 + rand.Float64()),
		targetY: float64(3 + rand.Intn(h/3)),
	}
}

func (r *rocket) explode() {
	count := 28 + rand.Intn(12)
	r.sparks = make([]spark, count)
	for i := range r.sparks {
		angle := float64(i) / float64(count) * 2 * math.Pi
		speed := 0.8 + rand.Float64()*2.0
		r.sparks[i] = spark{
			x:    r.x,
			y:    r.y,
			vx:   speed * math.Cos(angle),
			vy:   speed * math.Sin(angle) * 0.4,
			col:  confettiColors[rand.Intn(len(confettiColors))],
			ch:   sparkChars[rand.Intn(len(sparkChars))],
			life: 14 + rand.Intn(10),
		}
	}
	r.exploded = true
}

func stepRockets(rockets []rocket) []rocket {
	var alive []rocket
	for i := range rockets {
		r := rockets[i]
		if !r.exploded {
			r.y += r.vy
			if r.y <= r.targetY {
				r.y = r.targetY
				r.explode()
			}
			alive = append(alive, r)
			continue
		}
		anyAlive := false
		for j := range r.sparks {
			s := &r.sparks[j]
			if s.life <= 0 {
				continue
			}
			s.x += s.vx
			s.y += s.vy
			s.vy += 0.25
			s.life--
			anyAlive = true
		}
		if anyAlive {
			alive = append(alive, r)
		}
	}
	return alive
}

// ─── deploy lines ─────────────────────────────────────────────────────────────

var deployLines = []struct {
	delay int
	line  string
}{
	{0, sDim.Render("$ ") + sOrange.Render("helm upgrade --install birthday ./chart --namespace misha-bday")},
	{4, sDim.Render("Release \"birthday\" does not exist. Installing it now.")},
	{2, sGreen.Render("✓") + "  Namespace misha-bday created"},
	{3, sGreen.Render("✓") + "  ConfigMap/birthday-config applied"},
	{3, sGreen.Render("✓") + "  Deployment/birthday-app created"},
	{5, ""},
	{0, sDim.Render("$ ") + sBlue.Render("kubectl rollout status deployment/birthday-app -n misha-bday")},
	{4, sDim.Render("Waiting for deployment \"birthday-app\" rollout to finish: 0 of 3 updated replicas are available...")},
	{6, sDim.Render("Waiting for deployment \"birthday-app\" rollout to finish: 1 of 3 updated replicas are available...")},
	{5, sDim.Render("Waiting for deployment \"birthday-app\" rollout to finish: 2 of 3 updated replicas are available...")},
	{4, sGreen.Render("deployment \"birthday-app\" successfully rolled out")},
	{3, ""},
	{0, sDim.Render("$ ") + sBlue.Render("kubectl get pods -n misha-bday")},
	{3, sDim.Render("NAME                              READY   STATUS    RESTARTS   AGE")},
	{2, sGreen.Render("birthday-app-7d4f9b-xkcd9") + sDim.Render("         1/1     Running   0          3s")},
	{2, sGreen.Render("birthday-app-7d4f9b-m1sha") + sDim.Render("         1/1     Running   0          3s")},
	{2, sGreen.Render("birthday-app-7d4f9b-gonkv") + sDim.Render("         1/1     Running   0          3s")},
	{4, ""},
	{0, sDim.Render("$ ") + sBlue.Render("kubectl logs birthday-app-7d4f9b-m1sha -n misha-bday")},
	{3, sGold.Render("🎂  Happy Birthday, Mikhail Konkov!")},
	{3, sOrange.Render("   Tech Lead  |  DevOps  |  CI/CD  |  🟠 Orange Glasses")},
	{3, sGreen.Render("   Weight SLO: 81.7 kg → 70 kg  [▓▓▓▓▓▓░░░░░░] 58%  ✓")},
	{5, ""},
	{0, sGreen.Bold(true).Render("✓  RELEASE DEPLOYED SUCCESSFULLY  ✓")},
	{4, sDim.Render("Press Enter to continue →")},
}

// ─── jokes ────────────────────────────────────────────────────────────────────

type joke struct{ emoji, title, body string }

var jokes = []joke{
	{"📦", "Одноразка",
		"Одноразка — единственная штука в IT,\nкоторая заканчивается по-честному.\n\nНе «почти готово»,\nне «осталось немного».\n\nПросто закончилась — и всё.\n\nМиша это ценит."},
	{"🟠", "Мониторинг",
		"В Grafana всё красное.\nВ Kubernetes dashboard — жёлтое.\nВ Slack — срочное.\n\nМиша смотрит на всё это\nчерез оранжевые очки.\n\nСубъективно — норм."},
	{"🏁", "Бэк задеплоен",
		"Наконец-то.\n\nМиша закрыл 47 вкладок с логами.\nЗакрыл терминал с kubectl.\nПотянулся.\n\nПришёл фронтендер:\n— Слушай, а мы тоже\n  хотим в кубер..."},
	{"🎖️", "Техлид",
		"За день Мише задали вопросы:\n\n  — про архитектуру сервиса\n  — про упавший пайплайн\n  — про CORS (дважды)\n  — можно ли взять отгул в пятницу\n\nТолько последний\nбыл не по адресу."},
	{"☸️", "Всё в кубере",
		"Миша задеплоил в кубер:\n  бэкенд ✓\n  фронтенд ✓\n  мониторинг ✓\n  поздравление с ДР ✓\n\nСледующий квартал:\n  задеплоить саму команду\n  (в процессе обсуждения)"},
	{"🔄", "CI/CD",
		"Пайплайн упал на шаге deploy.\nМиша смотрит в логи.\nПайплайн упал на шаге build.\nМиша смотрит в логи.\nПайплайн прошёл.\n\nЭто и есть\nсчастливый конец."},
	{"🎛️", "Совещание",
		"— Нам нужно срочно!\n— Нам нужно качественно!\n— Нам нужно дёшево!\n\nМиша:\n— Выберите два.\n\nДолгая пауза.\n\n— ...А можно все три?\n— Нет.\n— А если очень попросить?\n— Нет."},
	{"🎂", "Релиз",
		"Каждый год — новый релиз.\n\n  + опыт\n  + кластеров стало больше\n  ~ одноразки: расход стабильный\n  ~ очки: оранжевые, без изменений\n  - терпение к плохим PR\n\nStatus: STABLE ✓"},
}

// ─── gifts ────────────────────────────────────────────────────────────────────

type gift struct {
	emoji, title, desc string
}

var gifts = []gift{
	{"🚬", "Бесконечный запас одноразок", "Жидкость больше никогда не закончится\nв самый неподходящий момент."},
	{"🛡️", "Год без инцидентов в проде", "Алертов нет. Slack молчит. Прод жив.\n365 дней подряд."},
	{"⚖️", "-11.7 кг прямо сейчас", "С 81.7 до 70. Мгновенно.\nSLO выполнен досрочно."},
	{"🤖", "Автодеплой фронта без вопросов", "Фронтендеры деплоятся сами.\nМиша просто наблюдает."},
	{"🏖️", "Отпуск без единого алерта", "Телефон молчит. PagerDuty спит.\nВсё само как-то работает."},
}

// ─── game ─────────────────────────────────────────────────────────────────────

const gameTimerMax = 100

type gameQuestion struct {
	pod, errName string
	opts         [3]string
	correct      int
}

var gameQuestions = []gameQuestion{
	{"frontend-app-7d4f9b", "CrashLoopBackoff",
		[3]string{"проверить логи", "выключить и включить", "уволить разработчика"}, 0},
	{"backend-svc-9x2kp", "OOMKilled",
		[3]string{"помолиться", "увеличить лимиты", "переименовать под"}, 1},
	{"worker-job-ab3cd", "ImagePullBackoff",
		[3]string{"перезагрузить ноутбук", "написать на Rust", "проверить registry"}, 2},
	{"nginx-ingress-5fg6", "Evicted",
		[3]string{"освободить диск", "добавить стикер на ноут", "игнорировать"}, 0},
	{"postgres-0", "Pending",
		[3]string{"переехать на SQLite", "проверить ресурсы", "перезагрузить роутер"}, 1},
	{"api-gw-3h7k", "Error",
		[3]string{"смотреть в логи", "написать в поддержку AWS", "удалить node_modules"}, 0},
	{"redis-master-1", "OOMKilled",
		[3]string{"добавить swap", "купить больше RAM", "увеличить limits"}, 2},
	{"birthday-app-m1sha", "Completed",
		[3]string{"ничего, это норм", "удалить под", "эскалировать в Slack"}, 0},
}

// ─── model ────────────────────────────────────────────────────────────────────

type model struct {
	screen screen
	w, h   int
	tick   int

	splashBlink bool

	deployStep    int
	deployVisible []string
	deployDone    bool

	jokeIdx   int
	jokeFlip  bool
	jokeFlipN int
	liked     map[int]bool

	giftIdx      int
	giftSelected int

	gameQ      int
	gameScore  int
	gameTimer  int
	gameFlash  string
	gameFlashN int
	gameDone   bool

	input textinput.Model
	wish  string

	particles []particle
	rockets   []rocket
}

func newModel() model {
	ti := textinput.New()
	ti.Placeholder = "Напиши пожелание Мише..."
	ti.CharLimit = 120
	ti.Focus()
	return model{
		liked:        make(map[int]bool),
		input:        ti,
		giftSelected: -1,
		gameTimer:    gameTimerMax,
	}
}

func (m model) Init() tea.Cmd {
	return tea.Batch(tickCmd(), deployTickCmd(), textinput.Blink)
}

// ─── update ───────────────────────────────────────────────────────────────────

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmds []tea.Cmd

	switch msg := msg.(type) {

	case tea.WindowSizeMsg:
		m.w, m.h = msg.Width, msg.Height

	case tickMsg:
		m.tick++
		m.splashBlink = m.tick%8 < 4

		if m.jokeFlip && m.jokeFlipN > 0 {
			m.jokeFlipN--
			if m.jokeFlipN == 0 {
				m.jokeFlip = false
			}
		}

		if m.screen == screenGame && !m.gameDone {
			if m.gameFlash != "" {
				m.gameFlashN--
				if m.gameFlashN <= 0 {
					m.gameFlash = ""
					m.gameQ++
					if m.gameQ >= len(gameQuestions) {
						m.gameDone = true
					} else {
						m.gameTimer = gameTimerMax
					}
				}
			} else {
				m.gameTimer--
				if m.gameTimer <= 0 {
					m.gameFlash = "timeout"
					m.gameFlashN = 10
				}
			}
		}

		if m.screen == screenFinal {
			m.particles = stepParticles(m.particles, m.h)
			if len(m.particles) < 50 && m.tick%4 == 0 {
				m.particles = append(m.particles, spawnParticles(m.w, 6)...)
			}
			m.rockets = stepRockets(m.rockets)
		}

		cmds = append(cmds, tickCmd())

	case deployStepMsg:
		if m.screen == screenDeploy && m.deployStep < len(deployLines) {
			e := &deployLines[m.deployStep]
			if e.delay > 0 {
				e.delay--
				cmds = append(cmds, deployTickCmd())
				return m, tea.Batch(cmds...)
			}
			m.deployVisible = append(m.deployVisible, e.line)
			m.deployStep++
			if m.deployStep >= len(deployLines) {
				m.deployDone = true
			} else {
				cmds = append(cmds, deployTickCmd())
			}
		}

	case tea.KeyMsg:
		switch m.screen {

		case screenSplash:
			switch msg.String() {
			case "enter", " ":
				m.screen = screenDeploy
				cmds = append(cmds, deployTickCmd())
			case "ctrl+c", "q":
				return m, tea.Quit
			}

		case screenDeploy:
			switch msg.String() {
			case "enter":
				if m.deployDone {
					m.screen = screenJokes
				}
			case "ctrl+c":
				return m, tea.Quit
			}

		case screenJokes:
			switch msg.String() {
			case "ctrl+c", "q":
				return m, tea.Quit
			case "right", "l", "tab", "d":
				m.jokeIdx = (m.jokeIdx + 1) % len(jokes)
				m.jokeFlip, m.jokeFlipN = true, 5
			case "left", "h", "a":
				m.jokeIdx = (m.jokeIdx - 1 + len(jokes)) % len(jokes)
				m.jokeFlip, m.jokeFlipN = true, 5
			case " ":
				m.liked[m.jokeIdx] = !m.liked[m.jokeIdx]
			case "enter":
				m.screen = screenGift
			}

		case screenGift:
			switch msg.String() {
			case "ctrl+c", "q":
				return m, tea.Quit
			case "up", "w":
				m.giftIdx = (m.giftIdx - 1 + len(gifts)) % len(gifts)
			case "down", "s":
				m.giftIdx = (m.giftIdx + 1) % len(gifts)
			case "enter":
				m.giftSelected = m.giftIdx
				m.screen = screenGame
				m.gameTimer = gameTimerMax
			case "esc":
				m.screen = screenJokes
			}

		case screenGame:
			switch msg.String() {
			case "ctrl+c", "q":
				return m, tea.Quit
			case "enter":
				if m.gameDone {
					m.screen = screenWish
				}
			case "1", "2", "3":
				if !m.gameDone && m.gameFlash == "" {
					ans := int(msg.String()[0] - '1')
					if ans == gameQuestions[m.gameQ].correct {
						m.gameFlash = "correct"
						m.gameScore++
					} else {
						m.gameFlash = "wrong"
					}
					m.gameFlashN = 9
				}
			}

		case screenWish:
			switch msg.String() {
			case "ctrl+c":
				return m, tea.Quit
			case "enter":
				if strings.TrimSpace(m.input.Value()) != "" {
					m.wish = m.input.Value()
					m.screen = screenFinal
					m.particles = spawnParticlesScattered(m.w, m.h, 80)
				}
			case "esc":
				m.screen = screenGame
			default:
				var tiCmd tea.Cmd
				m.input, tiCmd = m.input.Update(msg)
				cmds = append(cmds, tiCmd)
			}

		case screenFinal:
			switch msg.String() {
			case "ctrl+c", "q":
				return m, tea.Quit
			case " ":
				m.rockets = append(m.rockets, newRocket(m.w, m.h))
			}
		}
	}

	return m, tea.Batch(cmds...)
}

// ─── view ─────────────────────────────────────────────────────────────────────

func (m model) View() string {
	if m.w == 0 {
		return "Loading..."
	}
	switch m.screen {
	case screenSplash:
		return m.viewSplash()
	case screenDeploy:
		return m.viewDeploy()
	case screenJokes:
		return m.viewJokes()
	case screenGift:
		return m.viewGift()
	case screenGame:
		return m.viewGame()
	case screenWish:
		return m.viewWish()
	case screenFinal:
		return m.viewFinal()
	}
	return ""
}

// ─── splash ───────────────────────────────────────────────────────────────────

var splashArt = []string{
	` ██████╗ ██████╗     ██████╗  █████╗ ██╗   ██╗`,
	` ██╔══██╗╚════██╗   ██╔════╝ ██╔══██╗╚██╗ ██╔╝`,
	` ██║  ██║ █████╔╝   ██║  ███╗███████║ ╚████╔╝ `,
	` ██║  ██║ ╚═══██╗   ██║   ██║██╔══██║  ╚██╔╝  `,
	` ██████╔╝██████╔╝   ╚██████╔╝██║  ██║   ██║   `,
	` ╚═════╝ ╚═════╝     ╚═════╝ ╚═╝  ╚═╝   ╚═╝   `,
}

func (m model) viewSplash() string {
	var b strings.Builder
	b.WriteString(randomStripe(m.w, 2))

	artH := len(splashArt) + 10
	pad := (m.h - artH) / 2
	if pad < 1 {
		pad = 1
	}
	for i := 0; i < pad; i++ {
		b.WriteByte('\n')
	}

	pulse := []lipgloss.Color{cOrange, cGold, "#FFB347", cOrange}
	artStyle := lipgloss.NewStyle().Foreground(pulse[(m.tick/4)%len(pulse)]).Bold(true)
	for _, line := range splashArt {
		b.WriteString(center(artStyle.Render(line), m.w))
		b.WriteByte('\n')
	}
	b.WriteByte('\n')

	nameBox := lipgloss.NewStyle().Bold(true).Foreground(cWhite).
		BorderStyle(lipgloss.DoubleBorder()).BorderForeground(cOrange).Padding(0, 6).
		Render("🎂  Михаил Коньков  🎂")
	b.WriteString(center(nameBox, m.w))
	b.WriteString("\n\n")

	tags := lipgloss.JoinHorizontal(lipgloss.Center,
		tag("#00ADEF", "#000", " Go "), "  ",
		tag("#00C853", "#000", " DevOps "), "  ",
		tag("#FF8C00", "#000", " 🟠 Очки "), "  ",
		tag("#9B59B6", "#fff", " Tech Lead "),
	)
	b.WriteString(center(tags, m.w))
	b.WriteString("\n\n")

	prompt := sDim.Render("▶  Нажми Enter, чтобы задеплоить поздравление  ◀")
	if m.splashBlink {
		prompt = sGold.Render("▶  Нажми Enter, чтобы задеплоить поздравление  ◀")
	}
	b.WriteString(center(prompt, m.w))
	b.WriteByte('\n')
	b.WriteString(randomStripe(m.w, 2))
	return b.String()
}

// ─── deploy ───────────────────────────────────────────────────────────────────

func (m model) viewDeploy() string {
	boxW := clamp(m.w-8, 60, 100)
	var content strings.Builder
	content.WriteString(sOrange.Bold(true).Render("🚀  Deploying birthday-app...") + "\n\n")
	for _, line := range m.deployVisible {
		content.WriteString("  " + line + "\n")
	}
	if !m.deployDone && m.tick%6 < 3 {
		content.WriteString("  " + sGreen.Render("█"))
	}
	box := lipgloss.NewStyle().BorderStyle(lipgloss.RoundedBorder()).BorderForeground(cOrange).
		Padding(1, 2).Width(boxW).Render(content.String())

	var b strings.Builder
	vpad := (m.h - strings.Count(box, "\n") - 4) / 2
	for i := 0; i < vpad; i++ {
		b.WriteByte('\n')
	}
	b.WriteString(center(box, m.w))
	b.WriteByte('\n')
	if m.deployDone {
		b.WriteString(center(sGold.Render("  Enter → продолжить"), m.w))
	}
	return b.String()
}

// ─── jokes ────────────────────────────────────────────────────────────────────

func (m model) viewJokes() string {
	j := jokes[m.jokeIdx]
	cardW := clamp(m.w-12, 50, 72)

	flashColors := []lipgloss.Color{cWhite, cGold, cOrange, cOrange, cOrange}
	borderCol := cOrange
	if m.jokeFlip && m.jokeFlipN > 0 && m.jokeFlipN <= len(flashColors) {
		borderCol = flashColors[m.jokeFlipN-1]
	}

	likeStr := "  [ Space → ♡ ]"
	if m.liked[m.jokeIdx] {
		likeStr = lipgloss.NewStyle().Foreground(cPink).Bold(true).Render("  ♥  liked!")
	}

	card := lipgloss.NewStyle().BorderStyle(lipgloss.RoundedBorder()).BorderForeground(borderCol).
		Padding(1, 3).Width(cardW).
		Render(j.emoji + "  " + sBold.Foreground(cOrange).Render(j.title) + "\n\n" + sWhite.Render(j.body) + "\n\n" + likeStr)

	var dots strings.Builder
	likeCount := 0
	for i := range jokes {
		if i == m.jokeIdx {
			dots.WriteString(sOrange.Render("●"))
		} else {
			dots.WriteString(sDim.Render("○"))
		}
		if i < len(jokes)-1 {
			dots.WriteString(" ")
		}
	}
	for _, v := range m.liked {
		if v {
			likeCount++
		}
	}
	likeCounter := ""
	if likeCount > 0 {
		likeCounter = lipgloss.NewStyle().Foreground(cPink).Render(fmt.Sprintf("  ♥ %d", likeCount))
	}

	var b strings.Builder
	vpad := (m.h - strings.Count(card, "\n") - 8) / 2
	for i := 0; i < vpad; i++ {
		b.WriteByte('\n')
	}
	b.WriteString(center(sGold.Render("🎉  Поздравительный changelog  🎉"), m.w) + "\n\n")
	b.WriteString(center(card, m.w) + "\n\n")
	b.WriteString(center(sDim.Render("← A / D →  листать")+"   "+dots.String()+likeCounter, m.w) + "\n")
	b.WriteString(center(sDim.Render("Enter → дальше   q → выйти"), m.w) + "\n")
	return b.String()
}

// ─── gift ─────────────────────────────────────────────────────────────────────

func (m model) viewGift() string {
	boxW := clamp(m.w-12, 52, 68)
	innerW := boxW - 8

	var items strings.Builder
	for i, g := range gifts {
		if i == m.giftIdx {
			items.WriteString(sOrange.Bold(true).Render("▶") + "  " +
				lipgloss.NewStyle().Foreground(cGold).Bold(true).Render(g.emoji+"  "+g.title))
		} else {
			items.WriteString("   " + sDim.Render(g.emoji+"  "+g.title))
		}
		if i < len(gifts)-1 {
			items.WriteString("\n\n")
		}
	}

	sep := "\n\n" + sDim.Render(strings.Repeat("─", innerW)) + "\n"
	desc := lipgloss.NewStyle().Foreground(cWhite).Italic(true).Render(gifts[m.giftIdx].desc)

	box := lipgloss.NewStyle().BorderStyle(lipgloss.DoubleBorder()).BorderForeground(cGold).
		Padding(2, 3).Width(boxW).Render(items.String() + sep + desc)

	var b strings.Builder
	vpad := (m.h - strings.Count(box, "\n") - 6) / 2
	if vpad < 1 {
		vpad = 1
	}
	for i := 0; i < vpad; i++ {
		b.WriteByte('\n')
	}
	b.WriteString(center(sGold.Render("🎁  Выбери подарок для Миши"), m.w) + "\n\n")
	b.WriteString(center(box, m.w) + "\n\n")
	b.WriteString(center(sDim.Render("↑/↓  W/S — выбор   Enter — подарить   Esc — назад"), m.w) + "\n")
	return b.String()
}

// ─── game ─────────────────────────────────────────────────────────────────────

func (m model) viewGame() string {
	boxW := clamp(m.w-12, 52, 70)

	if m.gameDone {
		total := len(gameQuestions)
		emoji, comment := "😐", "Бывает."
		if m.gameScore >= total-1 {
			emoji, comment = "🏆", "Миша одобряет."
		} else if m.gameScore >= total/2 {
			emoji, comment = "👍", "Неплохо для не-DevOps."
		}
		result := lipgloss.NewStyle().BorderStyle(lipgloss.RoundedBorder()).BorderForeground(cGold).
			Padding(2, 4).Width(boxW).Align(lipgloss.Center).
			Render(sGold.Render("🔧  Игра окончена  "+emoji) + "\n\n" +
				sWhite.Render(fmt.Sprintf("Починено: %d / %d", m.gameScore, total)) + "\n\n" +
				sDim.Render(comment) + "\n\n" +
				sDim.Render("Enter → написать пожелание"))

		var b strings.Builder
		vpad := (m.h - strings.Count(result, "\n") - 2) / 2
		for i := 0; i < vpad; i++ {
			b.WriteByte('\n')
		}
		b.WriteString(center(result, m.w))
		return b.String()
	}

	q := gameQuestions[m.gameQ]

	// timer bar
	barW := boxW - 12
	filled := int(float64(barW) * float64(m.gameTimer) / float64(gameTimerMax))
	if filled < 0 {
		filled = 0
	}
	timerCol := cGreen
	if m.gameTimer < gameTimerMax/3 {
		timerCol = cRed
	} else if m.gameTimer < gameTimerMax*2/3 {
		timerCol = cGold
	}
	timerBar := lipgloss.NewStyle().Foreground(timerCol).Render(strings.Repeat("█", filled)) +
		sDim.Render(strings.Repeat("░", barW-filled))

	// options
	var opts strings.Builder
	for i, o := range q.opts {
		prefix := fmt.Sprintf("  %d.  ", i+1)
		switch {
		case m.gameFlash == "":
			opts.WriteString(sOrange.Render(prefix) + sWhite.Render(o))
		case i == q.correct:
			opts.WriteString(sGreen.Render(prefix) + sGreen.Render(o))
		default:
			opts.WriteString(sDim.Render(prefix) + sDim.Render(o))
		}
		if i < 2 {
			opts.WriteByte('\n')
		}
	}

	flashLine := ""
	switch m.gameFlash {
	case "correct":
		flashLine = "\n\n" + sGreen.Bold(true).Render("  ✓  Верно!")
	case "wrong":
		flashLine = "\n\n" + sRed.Bold(true).Render("  ✗  Неверно. Правильно: "+q.opts[q.correct])
	case "timeout":
		flashLine = "\n\n" + sRed.Bold(true).Render("  ⏱  Время вышло!")
	}

	progress := sDim.Render(fmt.Sprintf("%d / %d", m.gameQ+1, len(gameQuestions)))
	scoreStr := sGreen.Render(fmt.Sprintf("✓ %d", m.gameScore))

	content := lipgloss.NewStyle().BorderStyle(lipgloss.RoundedBorder()).BorderForeground(cOrange).
		Padding(1, 3).Width(boxW).
		Render(
			sOrange.Bold(true).Render("🔧  Почини под!") + "   " + progress + "   " + scoreStr + "\n\n" +
				sDim.Render("Pod:   ") + sBlue.Render(q.pod) + "\n" +
				sDim.Render("Error: ") + sRed.Bold(true).Render(q.errName) + "\n\n" +
				timerBar + "\n\n" +
				opts.String() +
				flashLine,
		)

	var b strings.Builder
	vpad := (m.h - strings.Count(content, "\n") - 4) / 2
	if vpad < 1 {
		vpad = 1
	}
	for i := 0; i < vpad; i++ {
		b.WriteByte('\n')
	}
	b.WriteString(center(content, m.w))
	if m.gameFlash == "" {
		b.WriteString("\n\n" + center(sDim.Render("Нажми 1, 2 или 3"), m.w))
	}
	return b.String()
}

// ─── wish ─────────────────────────────────────────────────────────────────────

func (m model) viewWish() string {
	boxW := clamp(m.w-12, 50, 72)
	m.input.Width = boxW - 8

	sep := sDim.Render(strings.Repeat("─", boxW-8))
	content := lipgloss.NewStyle().BorderStyle(lipgloss.RoundedBorder()).BorderForeground(cGold).
		Padding(2, 3).Width(boxW).
		Render(sGold.Render("💌  Напиши пожелание Мише") + "\n\n" +
			m.input.View() + "\n" + sep + "\n\n" +
			sDim.Render("Enter → отправить   Esc → назад"))

	var b strings.Builder
	vpad := (m.h - strings.Count(content, "\n") - 1) / 2
	if vpad < 1 {
		vpad = 1
	}
	for i := 0; i < vpad; i++ {
		b.WriteByte('\n')
	}
	b.WriteString(center(content, m.w))
	return b.String()
}

// ─── final ────────────────────────────────────────────────────────────────────

func (m model) viewFinal() string {
	// build grid
	grid := make([][]string, m.h)
	for i := range grid {
		row := make([]string, m.w)
		for j := range row {
			row[j] = " "
		}
		grid[i] = row
	}
	setCell := func(x, y int, s string) {
		if x >= 0 && x < m.w && y >= 0 && y < m.h {
			grid[y][x] = s
		}
	}

	for _, p := range m.particles {
		setCell(p.x, p.y, lipgloss.NewStyle().Foreground(p.col).Render(string(p.ch)))
	}
	for _, r := range m.rockets {
		if !r.exploded {
			setCell(int(r.x), int(r.y), lipgloss.NewStyle().Foreground(cGold).Bold(true).Render("|"))
		}
		for _, s := range r.sparks {
			if s.life > 0 {
				setCell(int(s.x), int(s.y), lipgloss.NewStyle().Foreground(s.col).Render(string(s.ch)))
			}
		}
	}

	// card
	cardW := clamp(m.w-12, 50, 72)
	likeCount := 0
	for _, v := range m.liked {
		if v {
			likeCount++
		}
	}
	giftLine := ""
	if m.giftSelected >= 0 {
		g := gifts[m.giftSelected]
		giftLine = "\n" + sOrange.Render("🎁  "+g.emoji+"  "+g.title)
	}
	scoreLine := ""
	if m.gameScore > 0 {
		scoreLine = "\n" + sDim.Render(fmt.Sprintf("🔧  Подов починено: %d / %d", m.gameScore, len(gameQuestions)))
	}
	wishLine := ""
	if m.wish != "" {
		wishLine = "\n\n" + lipgloss.NewStyle().Foreground(cGold).Italic(true).Render("\""+m.wish+"\"")
	}
	likeStr := ""
	if likeCount > 0 {
		likeStr = "\n" + lipgloss.NewStyle().Foreground(cPink).Render(fmt.Sprintf("♥  понравилось шуток: %d", likeCount))
	}
	card := lipgloss.NewStyle().Foreground(cWhite).
		BorderStyle(lipgloss.DoubleBorder()).BorderForeground(cOrange).
		Padding(1, 3).Width(cardW).Align(lipgloss.Center).
		Render(sOrange.Bold(true).Render("🎂  С днём рождения, Миша!  🎂") +
			giftLine + scoreLine + wishLine + likeStr +
			"\n\n" + sDim.Render("Space → фейерверк   q → выйти"))

	cardLines := strings.Split(card, "\n")
	startRow := (m.h - len(cardLines)) / 2
	colStart := (m.w - lipgloss.Width(card)) / 2
	if startRow < 0 {
		startRow = 0
	}
	if colStart < 0 {
		colStart = 0
	}

	var sb strings.Builder
	for row := 0; row < m.h; row++ {
		ci := row - startRow
		if ci >= 0 && ci < len(cardLines) {
			sb.WriteString(strings.Repeat(" ", colStart) + cardLines[ci] + "\n")
		} else {
			sb.WriteString(strings.Join(grid[row], "") + "\n")
		}
	}
	return sb.String()
}

// ─── helpers ──────────────────────────────────────────────────────────────────

func center(s string, w int) string {
	return lipgloss.PlaceHorizontal(w, lipgloss.Center, s)
}

func tag(bg, fg, text string) string {
	return lipgloss.NewStyle().Background(lipgloss.Color(bg)).Foreground(lipgloss.Color(fg)).
		Bold(true).Padding(0, 1).Render(text)
}

func randomStripe(w, rows int) string {
	if w == 0 {
		return ""
	}
	var sb strings.Builder
	for r := 0; r < rows; r++ {
		for i := 0; i < w; i++ {
			if rand.Intn(4) == 0 {
				ch := confettiChars[rand.Intn(len(confettiChars))]
				col := confettiColors[rand.Intn(len(confettiColors))]
				sb.WriteString(lipgloss.NewStyle().Foreground(col).Render(string(ch)))
			} else {
				sb.WriteByte(' ')
			}
		}
		sb.WriteByte('\n')
	}
	return sb.String()
}

func clamp(v, lo, hi int) int {
	if v < lo {
		return lo
	}
	if v > hi {
		return hi
	}
	return v
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}

// ─── main ─────────────────────────────────────────────────────────────────────

func main() {
	p := tea.NewProgram(newModel(), tea.WithAltScreen())
	if _, err := p.Run(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
