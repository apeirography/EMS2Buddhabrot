package main

import (
	"bufio"
	"bytes"
	"crypto/md5"
	"encoding/binary"
	"fmt"
	"image"
	"image/color"
	"image/png"
	"math"
	"math/cmplx"
	"math/rand"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"time"
)

/*****************************************************************************
 *  EMS2Buddhabrot generates Buddhabrot density graphs from EMS files.       *
 *  Copyright © 2020 Daïm Aggott-Hönsch                                      *
 *                                                                           *
 *  This program is free software: you can redistribute it and/or modify     *
 *  it under the terms of the GNU General Public License as published by     *
 *  the Free Software Foundation, either version 3 of the License, or        *
 *  (at your option) any later version.                                      *
 *                                                                           *
 *  This program is distributed in the hope that it will be useful,          *
 *  but WITHOUT ANY WARRANTY; without even the implied warranty of           *
 *  MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the            *
 *  GNU General Public License for more details.                             *
 *                                                                           *
 *  You should have received a copy of the GNU General Public License        *
 *  along with this program.  If not, see <https://www.gnu.org/licenses/>.   *
 *****************************************************************************/

func main() {

	fmt.Println("\nEMS2Buddhabrot v0.1 Copyright (C) 2020 Daïm Aggott-Hönsch. This program comes with ABSOLUTELY NO WARRANTY.")
	fmt.Println("This is free software, and you are welcome to redistribute it under the conditions specified by")
	fmt.Println("the GNU General Public License 3 (https://www.gnu.org/licenses/gpl-3.0).")
	
	if len(os.Args) < 2 {
		fmt.Println("\nEMS2Buddhabrot generates 2049x2049 images by default.  You can change the size of generated images by")
		fmt.Println("suffixing the filename an underscore and the image size.")
		fmt.Println("\nE.g.: " + filepath.Base(os.Args[0])[0:len(filepath.Base(os.Args[0]))-4] + "_" +"1500" + ".exe" + " <-- generate 1500x1500 images")
		fmt.Println("      " + filepath.Base(os.Args[0])[0:len(filepath.Base(os.Args[0]))-4] + "_" +"4000" + ".exe" + " <-- generate 4000x4000 images")
		
		fmt.Println("\nWARNING: No EMS files specified.  Drag and drop one or more EMS files on " + filepath.Base(os.Args[0]) + " to generate an image.")
		fmt.Println("\nGenerate EMS files with ectocopial Mandelbrot seeds using EMSMiner (https://github.com/apeirography/EMSMiner/releases).")
		
		time.Sleep(15 * time.Second)
	} else {
		fmt.Println("EMS2Buddhabrot generates 2049x2049 images by default.  You can change the size of generated images by")
		fmt.Println("suffixing the filename an underscore and the image size.")
		fmt.Println("\nE.g.: " + filepath.Base(os.Args[0])[0:len(filepath.Base(os.Args[0]))-4] + "_" +"1500" + ".exe" + " <-- generate 1500x1500 images")
		fmt.Println("      " + filepath.Base(os.Args[0])[0:len(filepath.Base(os.Args[0]))-4] + "_" +"4000" + ".exe" + " <-- generate 4000x4000 images")
		fmt.Println("")
		fmt.Println("Generate EMS files with ectocopial Mandelbrot seeds using EMSMiner (https://github.com/apeirography/EMSMiner/releases).")
		fmt.Println("")
	}
	
	imagesize := (1024 * 2) + 1
	
	soughtImagesize, err := strconv.Atoi(ExenameSubparts(".", 0, "_")[len(ExenameSubparts(".", 0, "_"))-1])
	if err == nil && soughtImagesize > 3 {
		imagesize = soughtImagesize
	}
	
	var inputFilenames []string
	for idx, _ := range os.Args {
		if idx > 0 {
			inputFilenames = append(inputFilenames, os.Args[idx])
		}
	}
	seeds := LoadEMSFile(inputFilenames[0])
	if len(inputFilenames) > 1 {
		for idx := 1; idx < len(inputFilenames); idx++ {
			moreSeeds := LoadEMSFile(inputFilenames[idx])
			if len(moreSeeds) > 0 {
				seeds = append(seeds, moreSeeds...)
			}
		}
	}

	fmt.Println(strconv.Itoa(len(seeds)) + " seeds loaded.")

	waypointMin, waypointMax := 1, 1000000000000

	startTime := time.Now()

	realMin, realMax := waypointMax, waypointMin

	width, height := imagesize, imagesize
	fmt.Println("Generating a " + strconv.Itoa(width) + "x" + strconv.Itoa(height) + " image.")
	
	minR, maxR := -2.00, 2.00
	minI, maxI := -2.00, 2.00
	delR := (maxR - minR) / float64(width)
	delI := (maxI - minI) / float64(height)

	pixels := make([]int, width*height)
	for idx := 0; idx < len(pixels); idx++ {
		pixels[idx] = 0
	}

	fmt.Print("Plotting density graph...")
	pixelmax := 0
	points := 0
	for sidx, seed := range seeds {
		depth := GetDepth(seed)
		seedIterationLimit := depth / 10 * 9 // For moderate depth seeds, the last 10% of waypoints will remain unplotted to reduce noise.
		if depth < 10 || depth > 9999 {
			seedIterationLimit = depth + 2
		}
		z := complex(0, 0)
		c := seed
		if rand.Intn(2) == 1 {
			c = complex(real(seed), imag(seed)*-1)
		}
		if sidx%1000 == 0 {
			fmt.Print(".")
		}
		if depth < realMin {
			realMin = depth
		}
		if depth > realMax {
			realMax = depth
		}
		for idx := 0; idx < seedIterationLimit && math.Abs(real(z)) <= 2 && math.Abs(imag(z)) <= 2; idx++ {
			if idx >= waypointMin && idx <= waypointMax {
				x, y := C2XY(z*complex(0, 1), minR, minI, delR, delI)
				pixels[y*width+x]++
				points++
				if pixels[y*width+x] > pixelmax {
					pixelmax = pixels[y*width+x]
				}
				if idx == waypointMax {
					break
				}
			}
			z = z*z + c
		}
	}
	fmt.Print(" done.\n")

	fmt.Print("Generating histogram to enhance contrast...")
	freqs := make(map[int]int)
	for x := 0; x < width; x++ {
		if x%10000 == 0 {
			fmt.Print(".")
		}
		for y := 0; y < height; y++ {
			freqs[pixels[y*width+x]]++
		}
	}

	cumfreqs := make(map[int]int)
	cumfreqs[1] = freqs[1]
	for i := 2; i < pixelmax; i++ {
		if i%10000 == 0 {
			fmt.Print(".")
		}
		cumfreqs[i] = cumfreqs[i-1] + freqs[i]
	}
	cumfreqsTotal := cumfreqs[len(cumfreqs)-1]

	fmt.Print(" done.\n")

	fmt.Print("Saving Buddhabrot...")

	totalseconds := int(math.Floor(time.Since(startTime).Seconds()))
	hours := totalseconds / 3600
	minutes := (totalseconds - (hours * 3600)) / 60
	seconds := totalseconds - (hours * 3600) - (minutes * 60)

	out, err := os.Create("Buddhabrot_of_"+strconv.Itoa(points)+"_waypoints_from_" + strconv.Itoa(len(seeds)) + "_seeds_with_depths_between_" + strconv.Itoa(realMin) + "-" + strconv.Itoa(realMax) + "_rendered_in_" + strconv.Itoa(hours) + "h" + strconv.Itoa(minutes) + "m" + strconv.Itoa(seconds) + "s.png")
	if err != nil {
		panic("\nCould not create file: " + err.Error())
	}
	img := image.NewRGBA(image.Rect(0, 0, width, height))
	for x := 0; x < width; x++ {
		for y := 0; y < height; y++ {
			//gs := pixels[y * width + x]
			gs := int(float64(cumfreqs[pixels[y*width+x]]) / float64(cumfreqsTotal) * 255)
			img.Set(x, y, color.RGBA{uint8((gs)), uint8(gs), uint8((gs)), 255})
		}
		if x % 32 == 0 {
			fmt.Print(".")
		}
	}
	png.Encode(out, img)
	out.Close()
	fmt.Print(" done.\n")
	time.Sleep(15 * time.Second)
}

func Exename() string {
	_, exename := filepath.Split(os.Args[0])
	return exename
}

func ExenameParts(separator string) []string {
	return strings.Split(Exename(), separator)
}

func ExenameSubparts(separator string, part int, subseparator string) []string {
	if len(ExenameParts(separator)) > part+1 {
		return strings.Split(ExenameParts(separator)[part], subseparator)
	}
	return []string{}
}

func GetDepth(c complex128) int {
	z := complex(0, 0)
	idx := 0
	for idx = 0; idx < 1000000000000 && cmplx.Abs(z) <= 2; idx++ {
		z = z*z+c
	}
	return idx
}

// Drawing functions

func C2XY(c complex128, minR, minI, delR, delI float64) (int, int) {
	x := int(math.Floor((real(c) - minR) / delR))
	y := int(math.Floor((imag(c) - minI) / delI))

	return x, y
}

// EMS file loading

func LoadEMSFile(filename string) seedpack {
	seeds := NewSeedpack(0)

	infilehandle, _ := os.OpenFile(filename, os.O_RDONLY, 0644)
	infile := bufio.NewReader(infilehandle)

	var tmp complex128
	var err error
	err = nil
	skip := 0
	signature := make([]byte, 32)
	err = binary.Read(infile, binary.LittleEndian, &signature)
	if string(signature) != "@DM.EMS{codex.apeirography.art} " {
		panic("File \"" + filename + "\" is not a valid EMS file.")
	}

	for err == nil {
		err = binary.Read(infile, binary.LittleEndian, &tmp)
		if skip > 0 {
			skip--
			continue
		}
		if err == nil {
			seeds = append(seeds, tmp)
		}
	}

	infilehandle.Close()

	return seeds
}

func SaveEMSFile(seeds seedpack, min, max int) {
	buf := new(bytes.Buffer)
	buf.Reset()

	seeds = seeds.Sort()

	for _, c := range seeds {
		binary.Write(buf, binary.LittleEndian, c)
	}

	md5 := md5.Sum(buf.Bytes())

	dir, _ := filepath.Abs(filepath.Dir(os.Args[0]))
	outfilename := filepath.Join(dir, strconv.Itoa(min)+"-"+strconv.Itoa(max)+"_"+fmt.Sprintf("%x", string(md5[:]))+".ems")
	outfile, err := os.OpenFile(outfilename, os.O_RDWR|os.O_CREATE, 0644)
	if err != nil {
		panic(err)
	}

	outfile = outfile

	buf.Reset()
	binary.Write(buf, binary.LittleEndian, []byte("@DM.EMS{codex.apeirography.art} "))
	for _, c := range seeds {
		binary.Write(buf, binary.LittleEndian, c)
	}
	outfile.Write(buf.Bytes())

	return
}

// Seedpack

type seedpack []complex128

func NewSeedpack(howmany int) seedpack {
	return seedpack(make([]complex128, howmany))
}

func (this seedpack) Sort() seedpack {
	sort.SliceStable(this, func(i, j int) bool {
		if real(this[i]) != real(this[j]) {
			return real(this[i]) < real(this[j])
		} else {
			return imag(this[i]) < imag(this[j])
		}
	})
	return this
}
