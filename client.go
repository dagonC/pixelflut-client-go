package main

import (
	"bufio"
	"fmt"
	"image"
	"image/color"
	"image/png"
	"math/rand"
	"net"
	"os"
	"strconv"
	"time"

	"github.com/nfnt/resize"
)

type pixel struct {
	x     int
	y     int
	color color.Color
}

type imageConfig struct {
	path  string
	width int
	x     int
	y     int
}

//creates a random distribution of pixels
//https://stackoverflow.com/questions/8697095/how-to-read-a-png-file-in-color-and-output-as-gray-scale-using-the-go-programmin
func buildRandomPixelCommandMap(imgCfg imageConfig, image image.Image, skipPixelModulo int) []string {
	//extracting pixel data from image
	fmt.Println("Extracting pixel data from image ...")
	imageBounds := image.Bounds()
	w, h := imageBounds.Max.X, imageBounds.Max.Y
	pixelSlice := make([]string, 1)

	//loop over image and fill pixel slice with ready made write pixel commands
	sliceIdx := 0
	for x := 0; x < w; x++ {
		for y := 0; y < h; y++ {
			if (sliceIdx)%skipPixelModulo == 0 {
				color := image.At(x, y)
				cmd := genPFWCFP(pixel{x: imgCfg.x + x, y: imgCfg.y + y, color: color})
				pixelSlice = append(pixelSlice, cmd)
			}
			sliceIdx = sliceIdx + 1
		}
	}

	fmt.Println("done.")

	//shuffle pixelSlice
	fmt.Println("Shuffling pixel data")
	rand.Seed(time.Now().UnixNano())
	rand.Shuffle(len(pixelSlice), func(i, j int) { pixelSlice[i], pixelSlice[j] = pixelSlice[j], pixelSlice[i] })
	fmt.Println("done.")

	return pixelSlice
}

func getImage(imgCfg imageConfig) image.Image {
	infile, err := os.Open(imgCfg.path)
	if err != nil {
		panic(err)
	}
	defer infile.Close()

	decodedImage, err := png.Decode(infile)
	if err != nil {
		panic(err)
	}

	resizedImage := resize.Resize(uint(imgCfg.width), calcHeight(imgCfg.width, decodedImage), decodedImage, resize.Lanczos3)

	return resizedImage
}

func calcHeight(desiredWidth int, decodedImage image.Image) uint {
	w, h := decodedImage.Bounds().Max.X, decodedImage.Bounds().Max.Y
	newHeight := uint((float64(h) / float64(w)) * float64(desiredWidth))
	fmt.Println(" - original  WxH: " + its(w) + "x" + its(h))
	fmt.Println(" - resized   WxH: " + its(desiredWidth) + "x" + its(int(newHeight)))
	return newHeight
}

func sendPixelCommandMapMulti(pcs []string, addr string, concurrentWorkers int) {
	fmt.Println("Sending pixels using " + its(concurrentWorkers) + " workers ...")
	chunkedPcs := chunkPixelSlices(pcs, concurrentWorkers)
	for i := 0; i < concurrentWorkers; i++ {
		go sendPixelWorker(i+1, concurrentWorkers, chunkedPcs[i], addr, 1)
	}
}

func chunkPixelSlices(pcs []string, numChunks int) [][]string {
	var chunked [][]string
	lenPcs := len(pcs)
	chunkSize := (lenPcs + numChunks - 1) / numChunks

	for i := 0; i < lenPcs; i += chunkSize {
		end := i + chunkSize

		if end > lenPcs {
			end = lenPcs
		}
		chunked = append(chunked, pcs[i:end])
	}
	return chunked
}

func sendPixelWorker(workerNumber int, maxWorkers int, wpcs []string, addr string, commandsPerConnection int) {
	wns := "worker " + its(workerNumber) + "/" + its(maxWorkers)
	fmt.Println(" >>" + wns + ": sending pixels ... ")
	numCommands := len(wpcs)
	var conn net.Conn
	var err error
	reconnect := true
	for true {
		for i := 0; i < numCommands; i++ {
			if reconnect == true {
				fmt.Println("connecting...")
				conn, err = net.Dial("tcp", addr)
				if err != nil {
					fmt.Println(err)
					continue
				} else {
					reconnect = false
				}
			}
			cmd := wpcs[i]
			if cmd != "" {
				_, err := sendPixel2(wpcs[i], conn)
				if err != nil {
					if conn != nil {
						conn.Close()
					}
					reconnect = true
				}
			}
		}
	}
	if conn != nil {
		conn.Close()
	}
}

func sendPixel2(spc string, conn net.Conn) (int, error) {
	w := bufio.NewWriter(conn)
	bytesWritten, err := w.WriteString(spc)
	if err != nil {
		fmt.Println(err)
		return 0, err
	}
	err = w.Flush()
	return bytesWritten, nil
}

func its(number int) string {
	return strconv.Itoa(number)
}

func iths(number uint32) string {
	return fmt.Sprintf("%02x", uint8(number))
}

func sti(str string) int {
	num, err := strconv.Atoi(str)
	if err != nil {
		panic(err)
	}
	return num
}

func genPFWCFP(p pixel) string {
	r, g, b, a := p.color.RGBA()
	cmd := ""
	if a != 0 {
		cmd = "PX " + its(p.x) + " " + its(p.y) + " " + iths(r) + iths(g) + iths(b) + iths(a) + "\n"
	}
	return cmd
}

func printUsage() {
	fmt.Println("usage:")
	fmt.Println(" client <IP> >PORT> <FILEPATH> <WIDTH> <X> <Y> <NUMBER OF WORKERS> [<SKIP PIXEL MODULO>]")
	fmt.Println("")
	fmt.Println("example:")
	fmt.Println(" client 94.45.232.48 1234 Logo_leiter.png 800 42 23 10 2")
}

func printHeader() {
	fmt.Println("K4CG Pixelflut Client - GO edition")
	fmt.Println("----------------------------------")
}

func printConfig(ip string, port string, imgCfg imageConfig) {
	fmt.Println("server:   " + ip + ":" + port)
	fmt.Println("image:    " + imgCfg.path)
	fmt.Println(" - x,y:   " + its(imgCfg.x) + ", " + its(imgCfg.y))
	fmt.Println(" - width: " + its(imgCfg.width))
}

func main() {
	printHeader()
	if len(os.Args) < 8 {
		printUsage()
		return
	}
	imgCfg := imageConfig{path: os.Args[3], width: sti(os.Args[4]), x: sti(os.Args[5]), y: sti(os.Args[6])}
	ip := os.Args[1]
	port := os.Args[2]
	numWorkers := sti(os.Args[7])
	pixelSkipModule := 1
	if len(os.Args) > 8 {
		pixelSkipModule = sti(os.Args[8])
	}
	printConfig(ip, port, imgCfg)

	image := getImage(imgCfg)
	pixelCommandMap := buildRandomPixelCommandMap(imgCfg, image, pixelSkipModule)
	sendPixelCommandMapMulti(pixelCommandMap, ip+":"+port, numWorkers)
	var input string
	fmt.Scanln(&input)
}
