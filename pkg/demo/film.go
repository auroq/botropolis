package demo

import (
	"encoding/json"
	"fmt"
	"image"
	_ "image/png"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"time"
)

// Shot is one piece of media made from a scenario: a still, or a clip
// when Record is set.
type Shot struct {
	Name        string   `yaml:"name"`
	Title       string   `yaml:"title"`
	Description string   `yaml:"description"`
	Window      string   `yaml:"window"`
	Keys        []string `yaml:"keys"`
	Hover       string   `yaml:"hover"`
	Projection  string   `yaml:"projection"`
	Detail      string   `yaml:"detail"`
	Signage     string   `yaml:"signage"`
	// Clock overrides the scenario's time of day for this shot.
	Clock string `yaml:"clock"`
	// Scale is the render scale; 0 follows the virtual display, which
	// is 1.
	Scale   float64    `yaml:"scale"`
	Record  *Recording `yaml:"record"`
	Formats []string   `yaml:"formats"`
}

// Recording is how long a clip runs and at what frame rate.
type Recording struct {
	Seconds float64 `yaml:"seconds"`
	FPS     int     `yaml:"fps"`
}

const (
	defaultClock  = "14:00"
	defaultWindow = "1920x1080"
	defaultFPS    = 30
	// noDaemon is a socket that cannot exist, which sends the city
	// straight to the staged files instead of whatever daemon the
	// machine is running against the real home.
	noDaemon = "/nonexistent/botropolis.sock"
)

var defaultFormats = []string{"mp4", "webm", "gif"}

type cityOption func(*[]string)

func withSocket(sock string) cityOption {
	return func(args *[]string) { (*args)[5] = sock }
}

// cityArgs is the botropolis command line for one shot: a screenshot to
// still, or frames into frames.
func cityArgs(shot Shot, home, still, frames string, opts ...cityOption) []string {
	args := []string{"city", "--headless", "--home", home, "--socket", noDaemon}
	for _, o := range opts {
		o(&args)
	}
	window := shot.Window
	if window == "" {
		window = defaultWindow
	}
	args = append(args, "--window", window)
	if len(shot.Keys) > 0 {
		args = append(args, "--keys", strings.Join(shot.Keys, ","))
	}
	for _, f := range []struct{ flag, value string }{
		{"--hover", shot.Hover}, {"--projection", shot.Projection}, {"--detail", shot.Detail}, {"--signage", shot.Signage},
	} {
		if f.value != "" {
			args = append(args, f.flag, f.value)
		}
	}
	if shot.Scale > 0 {
		args = append(args, "--render_scale", strconv.FormatFloat(shot.Scale, 'g', -1, 64))
	}
	if shot.Record != nil {
		return append(args, "--record", frames,
			"--seconds", strconv.FormatFloat(shot.Record.Seconds, 'g', -1, 64),
			"--fps", strconv.Itoa(fpsOf(shot.Record)))
	}
	return append(args, "--screenshot", still)
}

func fpsOf(r *Recording) int {
	if r.FPS > 0 {
		return r.FPS
	}
	return defaultFPS
}

// encodeArgs is the ffmpeg command line that turns recorded frames into
// one web format. H.264 in MP4 and VP9 in WebM between them play in
// every current browser; the GIF is for places that take no video,
// such as a README.
func encodeArgs(format, frames string, fps int, out string) ([]string, error) {
	args := []string{"-loglevel", "error", "-y", "-framerate", strconv.Itoa(fps), "-i", filepath.Join(frames, "frame-%05d.png")}
	switch format {
	case "mp4":
		args = append(args, "-c:v", "libx264", "-pix_fmt", "yuv420p", "-crf", "20", "-movflags", "+faststart")
	case "webm":
		args = append(args, "-c:v", "libvpx-vp9", "-pix_fmt", "yuv420p", "-crf", "34", "-b:v", "0")
	case "gif":
		args = append(args, "-vf", "fps=15,scale=960:-1:flags=lanczos,split[s0][s1];[s0]palettegen=max_colors=160[p];[s1][p]paletteuse=dither=bayer:bayer_scale=4")
	default:
		return nil, fmt.Errorf("format %q: want mp4, webm or gif", format)
	}
	return append(args, out), nil
}

// clockZone names a time zone in which now falls in the hour the clock
// asks for, so the city lights the scene for that time of day without
// anything in it being told to. Go reads TZ as a zone name only, so the
// zone is one of the Etc/GMT ones: Etc/GMT+N is N hours behind UTC, and
// they run from 12 behind to 14 ahead, which reaches every hour.
func clockZone(now time.Time, clock string) (string, error) {
	at, err := time.Parse("15:04", clock)
	if err != nil {
		return "", fmt.Errorf("clock %q: want HH:MM", clock)
	}
	ahead := ((at.Hour()-now.UTC().Hour())%24 + 24) % 24
	if ahead > 14 {
		ahead -= 24
	}
	switch {
	case ahead == 0:
		return "Etc/GMT", nil
	case ahead > 0:
		return fmt.Sprintf("Etc/GMT-%d", ahead), nil
	}
	return fmt.Sprintf("Etc/GMT+%d", -ahead), nil
}

// Manifest describes every file a film run made, for whatever builds the
// site to pick from without opening each one.
type Manifest struct {
	Generated time.Time `json:"generated"`
	Version   string    `json:"version"`
	Media     []Media   `json:"media"`
}

type Media struct {
	File        string  `json:"file"`
	Kind        string  `json:"kind"`
	Format      string  `json:"format"`
	Scenario    string  `json:"scenario"`
	Shot        string  `json:"shot"`
	Title       string  `json:"title"`
	Description string  `json:"description"`
	Width       int     `json:"width"`
	Height      int     `json:"height"`
	Seconds     float64 `json:"seconds,omitempty"`
	Poster      string  `json:"poster,omitempty"`
}

// Filmer stages each scenario afresh for every shot, so each one sees
// its sessions exactly as long ago as the scenario says, and runs the
// city against it with every home-shaped variable pointed into a
// scratch directory.
type Filmer struct {
	Corpus     Corpus
	Botropolis string
	FFmpeg     string
	Out        string
	Version    string
	Log        io.Writer
	Now        func() time.Time
}

// Film makes every shot of every scenario and writes the manifest.
func (f Filmer) Film(paths []string) (Manifest, error) {
	manifest := Manifest{Generated: f.Now().UTC(), Version: f.Version}
	paths, err := scenarioFiles(paths)
	if err != nil {
		return manifest, err
	}
	for _, path := range paths {
		scenario, err := LoadScenario(path)
		if err != nil {
			return manifest, err
		}
		if scenario.Name == "" {
			scenario.Name = strings.TrimSuffix(filepath.Base(path), filepath.Ext(path))
		}
		for _, shot := range scenario.Shots {
			media, err := f.shoot(scenario, shot)
			if err != nil {
				return manifest, fmt.Errorf("%s/%s: %w", scenario.Name, shot.Name, err)
			}
			manifest.Media = append(manifest.Media, media...)
		}
	}
	data, err := json.MarshalIndent(manifest, "", "  ")
	if err != nil {
		return manifest, err
	}
	return manifest, writeBytes(filepath.Join(f.Out, "manifest.json"), append(data, '\n'))
}

// scenarioFiles expands each directory to the scenario files in it, in
// name order, so a run films the same things in the same order.
func scenarioFiles(paths []string) ([]string, error) {
	var out []string
	for _, p := range paths {
		info, err := os.Stat(p)
		if err != nil {
			return nil, err
		}
		if !info.IsDir() {
			out = append(out, p)
			continue
		}
		found, err := filepath.Glob(filepath.Join(p, "*.yaml"))
		if err != nil {
			return nil, err
		}
		out = append(out, found...)
	}
	return out, nil
}

func (f Filmer) shoot(scenario Scenario, shot Shot) ([]Media, error) {
	work, err := os.MkdirTemp("", "botropolis-film-")
	if err != nil {
		return nil, err
	}
	defer func() { _ = os.RemoveAll(work) }()
	home := filepath.Join(work, "home")
	if err := os.MkdirAll(home, 0o700); err != nil {
		return nil, err
	}
	if err := Stage(f.Corpus, scenario, home, f.Now(), Sleepers(home, time.Hour)); err != nil {
		return nil, err
	}
	defer func() { _ = Unstage(home) }()
	dir := filepath.Join(f.Out, scenario.Name)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return nil, err
	}
	clock := firstOf(shot.Clock, scenario.Clock, defaultClock)
	tz, err := clockZone(f.Now(), clock)
	if err != nil {
		return nil, err
	}
	media := Media{Scenario: scenario.Name, Shot: shot.Name, Title: shot.Title, Description: shot.Description}
	fmt.Fprintf(f.Log, "filming %s/%s\n", scenario.Name, shot.Name)
	if shot.Record == nil {
		still := filepath.Join(dir, shot.Name+".png")
		if err := f.city(work, tz, cityArgs(shot, home, still, "")); err != nil {
			return nil, err
		}
		media.File, media.Kind, media.Format = rel(f.Out, still), "image", "png"
		media.Width, media.Height, err = size(still)
		return []Media{media}, err
	}
	frames := filepath.Join(work, "frames")
	if err := os.MkdirAll(frames, 0o700); err != nil {
		return nil, err
	}
	if err := f.city(work, tz, cityArgs(shot, home, "", frames)); err != nil {
		return nil, err
	}
	poster := filepath.Join(dir, shot.Name+".png")
	if err := posterFrom(frames, poster); err != nil {
		return nil, err
	}
	width, height, err := size(poster)
	if err != nil {
		return nil, err
	}
	formats := shot.Formats
	if len(formats) == 0 {
		formats = defaultFormats
	}
	var out []Media
	for _, format := range formats {
		file := filepath.Join(dir, shot.Name+"."+format)
		args, err := encodeArgs(format, frames, fpsOf(shot.Record), file)
		if err != nil {
			return nil, err
		}
		if err := run(exec.Command(f.FFmpeg, args...), f.Log); err != nil {
			return nil, fmt.Errorf("ffmpeg %s: %w", format, err)
		}
		m := media
		m.File, m.Kind, m.Format, m.Poster = rel(f.Out, file), "video", format, rel(f.Out, poster)
		m.Width, m.Height, m.Seconds = width, height, shot.Record.Seconds
		if format == "gif" {
			m.Kind = "animation"
			m.Width, m.Height = 960, height*960/width
		}
		out = append(out, m)
	}
	return out, nil
}

// city runs botropolis with HOME, the XDG directories and the config
// all inside the work dir: the city's layout file, its settings and
// anything else it reaches for by home directory land there, never in
// the real ones.
func (f Filmer) city(work, tz string, args []string) error {
	cmd := exec.Command(f.Botropolis, args...)
	env := []string{
		"HOME=" + filepath.Join(work, "home"),
		"XDG_STATE_HOME=" + filepath.Join(work, "state"),
		"XDG_CONFIG_HOME=" + filepath.Join(work, "config"),
		"XDG_CACHE_HOME=" + filepath.Join(work, "cache"),
		"XDG_RUNTIME_DIR=" + filepath.Join(work, "run"),
		"TZ=" + tz,
	}
	for _, kv := range os.Environ() {
		key, _, _ := strings.Cut(kv, "=")
		switch key {
		case "PATH", "LANG", "LC_ALL", "TMPDIR":
			env = append(env, kv)
		}
	}
	cmd.Env = env
	return run(cmd, f.Log)
}

func firstOf(values ...string) string {
	for _, v := range values {
		if v != "" {
			return v
		}
	}
	return ""
}

func run(cmd *exec.Cmd, log io.Writer) error {
	cmd.Stdout, cmd.Stderr = log, log
	return cmd.Run()
}

// posterFrom keeps the frame a third of the way in as the clip's
// still: the first frame is the city before its fit has settled.
func posterFrom(frames, poster string) error {
	all, err := filepath.Glob(filepath.Join(frames, "frame-*.png"))
	if err != nil || len(all) == 0 {
		return fmt.Errorf("no frames recorded in %s", frames)
	}
	data, err := os.ReadFile(all[len(all)/3])
	if err != nil {
		return err
	}
	return os.WriteFile(poster, data, 0o644)
}

func size(path string) (int, int, error) {
	file, err := os.Open(path)
	if err != nil {
		return 0, 0, err
	}
	defer func() { _ = file.Close() }()
	config, _, err := image.DecodeConfig(file)
	return config.Width, config.Height, err
}

func rel(base, path string) string {
	r, err := filepath.Rel(base, path)
	if err != nil {
		return path
	}
	return filepath.ToSlash(r)
}
