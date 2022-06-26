package main

import (
	"bytes"
	_ "embed"
	"fmt"
	"image"
	"image/color"
	_ "image/png"
	"log"
	"strings"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/examples/resources/fonts"
	"github.com/hajimehoshi/ebiten/v2/inpututil"
	"github.com/hajimehoshi/ebiten/v2/text"
	"golang.org/x/image/font"
	"golang.org/x/image/font/opentype"
)

const (
	screenWidth  = 300
	screenHeight = 300

	frameOX     = 0
	frameOY     = 32
	frameWidth  = 150
	frameHeight = 150
	frameNum    = 8
	fontSize    = 10

	// gmae modes
	modeTitle = 0
	modeGame  = 1
)

var (
	dinosaur1Img *ebiten.Image
	arcadeFont   font.Face
)

//go:embed resources/images/dinosaur_01.png
var byteDinosaur1Img []byte

// ebiten.Game interface を満たす型がEbitenには必要なので、
// この Game 構造体に Update, Draw, Layout 関数を持たせます。
type Game struct {
	count      int
	mode       int
	score      int
	hiscore    int
	velocity   int
	charge     int
	prevKey    int
	currentKey int

	keys []ebiten.Key
}

// 構造体の初期化を行なっています。
func (g *Game) init() *Game {
	g.hiscore = g.score
	g.count = 0
	g.score = 0
	g.velocity = 0

	return g
}

// Update関数は、画面のリフレッシュレートに関わらず
// 常に毎秒60回呼ばれます（既定値）。
// 描画ではなく更新処理を行うことが推奨されます。
func (g *Game) Update() error {
	g.count++

	switch g.mode {
	case modeTitle:
		if g.isKeyJustPressed(ebiten.KeySpace) {
			g.mode = modeGame
		}
	case modeGame:
		if g.isKeyJustPressed(ebiten.KeySpace) {
			g.mode = modeTitle
		}
	}

	// キー入力をフレーム毎に受付
	g.keys = inpututil.AppendPressedKeys(g.keys[:0])

	return nil
}

// スペースキーが押されたかを判定しています
func (g *Game) isKeyJustPressed(key ebiten.Key) bool {
	if inpututil.IsKeyJustPressed(key) {
		return true
	}
	return false
}

// Draw関数は、画面のリフレッシュレートと同期して呼ばれます（既定値）。
// 描画処理のみを行うことが推奨されます。ここで状態の変更を行うといろいろ事故ります。
func (g *Game) Draw(screen *ebiten.Image) {
	screen.Fill(color.White)

	// キーボード入力を表示させるを
	keyStrs := []string{}
	for _, p := range g.keys {
		keyStrs = append(keyStrs, p.String())
	}

	text.Draw(screen, fmt.Sprintf("Hiscore: %d", g.hiscore), arcadeFont, 20, 20, color.Black)
	text.Draw(screen, fmt.Sprintf("score: %d", g.score), arcadeFont, 20, 35, color.Black)
	text.Draw(screen, fmt.Sprintf("mode: %d", g.mode), arcadeFont, 20, 50, color.Black)
	text.Draw(screen, fmt.Sprintf("Keys: %s", strings.Join(keyStrs, ", ")), arcadeFont, 20, 65, color.Black)
	text.Draw(screen, fmt.Sprintf("velocity: %d", g.velocity), arcadeFont, 20, 80, color.Black)
	text.Draw(screen, fmt.Sprintf("charge: %d", g.charge), arcadeFont, 20, 95, color.Black)
	text.Draw(screen, fmt.Sprintf("prevKey: %d", g.prevKey), arcadeFont, 20, 110, color.Black)
	text.Draw(screen, fmt.Sprintf("currentKey: %d", g.currentKey), arcadeFont, 20, 125, color.Black)
	text.Draw(screen, fmt.Sprintf("g.count: %d", g.count), arcadeFont, 20, 140, color.Black)

	// ebitenで画像を表示に関わるオプション設定をします
	option := &ebiten.DrawImageOptions{}
	option.GeoM.Translate(-float64(frameWidth)/2, -float64(frameHeight)/2)
	option.GeoM.Translate(screenWidth/2, screenHeight/2)

	// 長方形画像をスライドして切り出すことで、表示させたい画像を抽出する。
	// Rect(x0, y0, x1, y1)で(x0, y0),(x1 ,y1 )の範囲を切り出す
	// x軸はslideX ~ slideX + frmaWidth の範囲。iの直で可変。
	// y軸はslideY ~ slideY + frmaeHeight で固定値。
	i := (g.count / 5) % frameNum
	slideX, slideY := frameOX+i*frameWidth, frameOY
	rectAngle := image.Rect(slideX, slideY, slideX+frameWidth, slideY+frameHeight)
	screen.DrawImage(dinosaur1Img.SubImage(rectAngle).(*ebiten.Image), option)
}

// Layout関数は、ウィンドウのリサイズの挙動を決定します。画面サイズを返すのが無難だが適宜調整してください。
func (g *Game) Layout(outsideWidth, outsideHeight int) (int, int) {
	return screenWidth, screenHeight
}

// init関数はパッケージの初期化に使われる特殊な関数で、main関数が呼ばれる前に実行されます。
// ここではとりあえず画像ファイル、フォントデータを読み込むのに利用しています。
func init() {
	img, _, err := image.Decode(bytes.NewReader(byteDinosaur1Img))
	if err != nil {
		log.Fatal(err)
	}
	dinosaur1Img = ebiten.NewImageFromImage(img)

	tt, err := opentype.Parse(fonts.PressStart2P_ttf)
	if err != nil {
		log.Fatal(err)
	}
	const dpi = 72
	arcadeFont, err = opentype.NewFace(tt, &opentype.FaceOptions{
		Size:    fontSize,
		DPI:     dpi,
		Hinting: font.HintingFull,
	})
}

// main関数から全ての処理が動き出します。ただしinit関数などの特殊な関数は除きます。
func main() {
	if err := _main(); err != nil {
		panic(err)
	}
}

func _main() error {
	g, err := newGame()
	if err != nil {
		return err
	}

	// ウィンドウズサイズとウィンドウ上部の表示タイトルを指定します。
	ebiten.SetWindowTitle("Animation (Ebiten Demo)")
	ebiten.SetWindowSize(screenWidth*2, screenHeight*2)
	return ebiten.RunGame(g)
}

func newGame() (*Game, error) {
	// type struct
	g := &Game{}
	g.init()
	return g, nil
}
