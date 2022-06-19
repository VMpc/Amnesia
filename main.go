package main

import (
	"encoding/base64"
	"flag"
	"fmt"
	"image/color"
	"os"
	"runtime"
	"strings"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/widget"
	"golang.org/x/crypto/argon2"
)

type params struct {
	memory      uint32
	iterations  uint32
	parallelism uint8
	saltLength  uint32
	keyLength   uint32
}

func main() {
	p := &params{
		memory:      128 * 1024,
		iterations:  1,
		parallelism: uint8(runtime.NumCPU() * 2),
		saltLength:  16,
		keyLength:   32,
	}

	if len(os.Args) >= 2 {
		masterPass := flag.String("master", "NONE", "The main password can stay the same across sites, just remember you need to not have AMNESIA when you need to recover the password")
		secretKey := flag.String("secret", "None", "The secret key can be anything you want (Aslong as it isn't re-used across sites), just remember you need to not have AMNESIA when you need to recover the password")
		memory := flag.Int("memory", 128*1024, "Changes the memory (in KB) param of argon2id")
		iterations := flag.Int("iterations", 1, "Changes the iterations param of argon2id")
		parallelism := flag.Int("parallelism", runtime.NumCPU()*2, "Changes the parallelism param of argon2id")
		saltLength := flag.Int("saltlength", 16, "Changes the salt length param of argon2id")
		keyLength := flag.Int("keyLength", 32, "Changes the key length param of argon2id")

		flag.Parse()

		if len(*masterPass) == 0 || len(*secretKey) == 0 {
			fmt.Println("Both the master password and secret key need to be provided")
			os.Exit(1)
		}

		p.memory = uint32(*memory)
		p.iterations = uint32(*iterations)
		p.parallelism = uint8(*parallelism)
		p.saltLength = uint32(*saltLength)
		p.keyLength = uint32(*keyLength)

		hash, err := genPassword([]byte(*masterPass), []byte(*secretKey), p)
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}

		fmt.Println(hash)

		return
	}

	app := app.NewWithID("com.github.Not_Cyrus.Amnesia")
	win := app.NewWindow("Amnesia")
	win.Resize(fyne.NewSize(550, 350))
	win.CenterOnScreen()
	win.SetFixedSize(true)

	mainPass := widget.NewPasswordEntry()
	secretKey := widget.NewPasswordEntry()
	outLabel := genLabel("", true)
	tipText := canvas.NewText("Tips", color.White)
	tipText.TextStyle.Bold = true

	tipTitle := container.New(layout.NewCenterLayout(), tipText)

	mainPass.SetPlaceHolder("Enter your main password...")
	secretKey.SetPlaceHolder("Enter your secret key...")

	win.SetContent(container.NewVBox(
		mainPass,
		secretKey,
		outLabel,

		widget.NewButton("Generate Password", func() {
			if len(mainPass.Text) == 0 || len(secretKey.Text) == 0 {
				app.SendNotification(fyne.NewNotification("Amnesia", "You need to enter your main password and your secret key"))
				return
			}

			hash, err := genPassword([]byte(mainPass.Text), []byte(secretKey.Text), p)
			if err != nil {
				app.SendNotification(fyne.NewNotification("Amnesia", "Failed to generate a password: "+err.Error()))
				return
			}

			outLabel.Show()
			app.SendNotification(fyne.NewNotification("Amnesia", "Password has been generated"))
			outLabel.SetText("Generated Password is: " + hash)
		}),

		widget.NewButton("Copy to Clipboard", func() {
			if !strings.Contains(outLabel.Text, "Generated Password is: ") {
				app.SendNotification(fyne.NewNotification("Amnesia", "Could not copy anything to clipboard"))
				return
			}

			win.Clipboard().SetContent(strings.Replace(outLabel.Text, "Generated Password is: ", "", 1))
			app.SendNotification(fyne.NewNotification("Amnesia", "Password has been copied to clipboard"))
		}),

		widget.NewButton("Clear Clipboard", func() {
			win.Clipboard().SetContent("")
			app.SendNotification(fyne.NewNotification("Amnesia", "Clipboard has been cleared"))

			go func() {
				time.Sleep(5 * time.Second)
				outLabel.Hide()
			}()
		}),

		widget.NewSeparator(),
		tipTitle,
		widget.NewSeparator(),
		genLabel("The main password should be something that you can remember as there will be no way of recovering the password otherwise", false),
		widget.NewSeparator(),
		genLabel("The secret key could be something like Username@example.com or a codephrase do NOT share these between websites, it ruins the point of this program", false),
	))

	win.ShowAndRun()
}

func genLabel(text string, hide bool) (Label *widget.Label) {
	Label = widget.NewLabel(text)
	Label.Wrapping = 3
	if hide {
		Label.Hide()
	}

	return
}

func genPassword(password, salt []byte, p *params) (encodedHash string, err error) {

	hash := argon2.IDKey(password, salt, p.iterations, p.memory, p.parallelism, p.keyLength)

	b64Salt := base64.RawStdEncoding.EncodeToString(salt)
	b64Hash := base64.RawStdEncoding.EncodeToString(hash)

	encodedHash = fmt.Sprintf("%s%s", b64Hash, b64Salt)

	return encodedHash, nil
}
