package main

import (
	"bytes"
	_ "embed"
	"fmt"
	"image"
	"image/color"
	_ "image/png"
	"log"
	"math"

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

	frameWidth  = 150
	frameHeight = 150
	frameNum    = 8
	fontSize    = 10
	coefficient = 0.4

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
	count    int
	mode     int
	score    int
	hiscore  int
	acceleration int
	charge   int

	prevKey    ebiten.Key
	currentKey ebiten.Key
	keys       []ebiten.Key
}

// 構造体の初期化を行なっています。
func (g *Game) init() *Game {
	g.hiscore = g.score
	g.count = 0
	g.score = 0
	g.acceleration = 0

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
		if g.isKeyJustPressed(ebiten.KeyArrowLeft) {
			if g.prevKey == ebiten.KeyArrowRight {
				g.acceleration += 1
			} else if g.prevKey == ebiten.KeyArrowLeft {
				g.acceleration -= 1
			}
			g.prevKey = ebiten.KeyArrowLeft
			g.currentKey = ebiten.KeyArrowLeft
		}
		if g.isKeyJustPressed(ebiten.KeyArrowRight) {
			if g.prevKey == ebiten.KeyArrowLeft {
				g.acceleration += 1
			} else if g.prevKey == ebiten.KeyArrowRight {
				g.acceleration -= 1
			}
			g.prevKey = ebiten.KeyArrowRight
			g.currentKey = ebiten.KeyArrowRight
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

	text.Draw(screen, fmt.Sprintf("mode: %d", g.mode), arcadeFont, 20, 20, color.Black)
	text.Draw(screen, fmt.Sprintf("acceleration: %d", g.acceleration), arcadeFont, 20, 30, color.Black)
	text.Draw(screen, fmt.Sprintf("charge: %d", g.charge), arcadeFont, 20, 40, color.Black)
	text.Draw(screen, fmt.Sprintf("g.count: %d", (g.count%360)), arcadeFont, 20, 50, color.Black)

	// ebitenで画像を表示に関わるオプション設定をします
	option := &ebiten.DrawImageOptions{}

	// 画像の中心をスクリーンの左上に移動させる
	// ジオメトリマトリックス（回転や移動の処理）が適用される時の
	// 原点が画面の左上だから、加工のために中心に配置される画像を一旦原点に移動させる
	option.GeoM.Translate(-float64(screenWidth)/2, -float64(screenHeight)/2)

	// 構造体の状態を元に回転角度を算出する
	option.GeoM.Rotate(float64(float64((g.count * g.acceleration)%360) * 2 * math.Pi / 360))

	// 画像を拡大/縮小する
	option.GeoM.Scale(coefficient, coefficient)

	// 画像を好きな位置に移動させる
	// 今回は画像をスクリーンの中心に持ってくる
	option.GeoM.Translate(screenWidth/2, screenHeight/2)

	// オプションを元に画像を描画する
	screen.DrawImage(dinosaur1Img, option)
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
