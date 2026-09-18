package render

import (
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/hajimehoshi/ebiten/v2"

	"github.com/auroq/botropolis/pkg/city"
	"github.com/auroq/botropolis/pkg/state"
	"github.com/auroq/botropolis/pkg/ui"
)

// bindings is every key the city answers to, in the order the help lists
// them; the footer shows the few that matter most.
var bindings = []ui.Key{
	{Key: "drag", Action: "pan"},
	{Key: "arrows", Action: "pan"},
	{Key: "wheel", Action: "zoom"},
	{Key: "+ / -", Action: "zoom"},
	{Key: "click", Action: "attach"},
	{Key: "enter", Action: "attach the selection"},
	{Key: "tab", Action: "next needs-you"},
	{Key: "c", Action: "new session here"},
	{Key: "d d", Action: "demolish the selection"},
	{Key: "f", Action: "fit"},
	{Key: "0", Action: "reset the view"},
	{Key: "r", Action: "turn the camera"},
	{Key: "n", Action: "night, day, auto"},
	{Key: "[ ]", Action: "scrub the clock an hour"},
	{Key: "s", Action: "settings"},
	{Key: "x", Action: "power breakdown (also click the plant)"},
	{Key: "t", Action: "timeline; enter jumps to the session"},
	{Key: "b", Action: "sidebar"},
	{Key: "h", Action: "hide the UI"},
	{Key: "p", Action: "save a screenshot"},
	{Key: "?", Action: "this help"},
	{Key: "q", Action: "quit"},
}

var footerKeys = []ui.Key{
	{Key: "drag", Action: "pan"},
	{Key: "wheel", Action: "zoom"},
	{Key: "click", Action: "attach"},
	{Key: "tab", Action: "next needs-you"},
	{Key: "b", Action: "sidebar"},
	{Key: "f", Action: "fit"},
	{Key: "?", Action: "help"},
	{Key: "q", Action: "quit"},
}

const keyPanStep = 12.0

// handleKeys answers the keyboard for one tick.
func (g *Game) handleKeys() error {
	just := g.just
	if g.settingsOpen {
		g.settingsOpen = g.settingsKeys()
		return nil
	}
	if g.breakdown {
		g.breakdown = g.breakdownKeys()
		return nil
	}
	if g.timeline {
		g.timeline = g.timelineKeys()
		return nil
	}
	if just(ebiten.KeyT) {
		g.timeline, g.away, g.cursor = true, nil, 0
		return nil
	}
	if just(ebiten.KeyX) {
		g.breakdown, g.window = true, city.LastDay
		return nil
	}
	if just(ebiten.KeyS) {
		g.settingsOpen = true
		return nil
	}
	if just(ebiten.KeyEscape) {
		if g.help {
			g.help = false
			return nil
		}
		return g.quit()
	}
	if just(ebiten.KeyQ) {
		return g.quit()
	}
	if just(ebiten.KeySlash) || just(ebiten.KeyF1) {
		g.help = !g.help
	}
	if just(ebiten.KeyH) {
		g.hidden = !g.hidden
	}
	if just(ebiten.KeyB) {
		g.sidebar = !g.sidebar
	}
	if just(ebiten.KeyP) {
		g.snap = screenshotPath(time.Now())
	}
	if just(ebiten.KeyF) || just(ebiten.KeyDigit0) || just(ebiten.KeyKP0) {
		g.scene.Fit()
	}
	if just(ebiten.KeyTab) {
		g.jump(state.NeedsYou)
	}
	if just(ebiten.KeyEnter) {
		g.act(g.scene.Activate())
	}
	if just(ebiten.KeyBracketLeft) || just(ebiten.KeyBracketRight) {
		by := time.Hour
		if just(ebiten.KeyBracketLeft) {
			by = -time.Hour
		}
		shift := g.scene.Scrub(by)
		g.SetStatus(fmt.Sprintf("clock %s (%+.0fh; n to go live)", g.scene.Clock().Format("15:04"), shift.Hours()))
	}
	if just(ebiten.KeyR) {
		g.SetStatus(fmt.Sprintf("heading %d°", g.scene.Turn()))
	}
	if just(ebiten.KeyC) {
		action, note := g.scene.NewHere()
		g.SetStatus(note)
		g.act(action)
	}
	if just(ebiten.KeyN) {
		switch g.scene.CycleLight() {
		case city.LightNight:
			g.SetStatus("light: night (n again for day)")
		case city.LightDay:
			g.SetStatus("light: day (n again to follow the clock)")
		default:
			g.scene.Scrub(0)
			g.SetStatus("light: live")
		}
	}
	if just(ebiten.KeyD) || just(ebiten.KeyDelete) {
		action, note := g.scene.Demolish(time.Now())
		g.SetStatus(note)
		g.act(action)
	}
	var pan city.Point
	if ebiten.IsKeyPressed(ebiten.KeyArrowLeft) {
		pan.X += keyPanStep
	}
	if ebiten.IsKeyPressed(ebiten.KeyArrowRight) {
		pan.X -= keyPanStep
	}
	if ebiten.IsKeyPressed(ebiten.KeyArrowUp) {
		pan.Y += keyPanStep
	}
	if ebiten.IsKeyPressed(ebiten.KeyArrowDown) {
		pan.Y -= keyPanStep
	}
	if pan != (city.Point{}) {
		g.scene.Pan(pan.Scale(g.theme.Scale))
	}
	centre := g.scene.Size().Scale(0.5)
	if just(ebiten.KeyEqual) || just(ebiten.KeyKPAdd) {
		g.scene.Wheel(centre, 1)
	}
	if just(ebiten.KeyMinus) || just(ebiten.KeyKPSubtract) {
		g.scene.Wheel(centre, -1)
	}
	return nil
}

func (g *Game) quit() error {
	if g.saveState != nil {
		g.saveState(g.scene.Layout())
	}
	return ebiten.Termination
}

// act hands an action to the actor and reports a failure in the footer.
func (g *Game) act(action city.Action) {
	if action.Kind == city.ActionNone || g.actor == nil {
		return
	}
	if err := g.actor.Do(action); err != nil {
		g.SetStatus(err.Error())
	}
}

// screenshotPath is where p saves a frame: the Pictures directory when
// there is one, else home, stamped to the second.
func screenshotPath(now time.Time) string {
	dir, err := os.UserHomeDir()
	if err != nil {
		dir = "."
	}
	if pictures := filepath.Join(dir, "Pictures"); isDir(pictures) {
		dir = pictures
	}
	return filepath.Join(dir, fmt.Sprintf("botropolis-%s.png", now.Format("2006-01-02-150405")))
}

func isDir(path string) bool {
	info, err := os.Stat(path)
	return err == nil && info.IsDir()
}
