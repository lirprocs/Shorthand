package decrypt

import (
	"image"
	"image/color"
	"math/rand"
	alphabet "shorthand/Alphabet"
	encrypt "shorthand/Encryption"
	"strconv"
	"strings"
	"sync"
)

func GetInfo(file image.Image, pol *map[int][]int) uint64 {
	x, y := encrypt.X1, encrypt.Y1
	text := "00000000"
	for i := 0; i < 2; i++ {
		oldColor := file.At(x, y).(color.RGBA)
		r := encrypt.UintToBinary(oldColor.R, 7)
		g := encrypt.UintToBinary(oldColor.G, 7)
		b := encrypt.UintToBinary(oldColor.B, 7)
		text = text + r[4:] + g[4:] + b[4:]
		x, y = x+1, y+1
		(*pol)[y] = append((*pol)[y], x)
	}
	//fmt.Println(text)
	lenn, _ := strconv.ParseUint(text, 2, 64)
	return lenn
}

func GetBin(wg *sync.WaitGroup, file image.Image, x, y int, i uint64, list *[]uint8) {
	defer wg.Done()
	i = i - 1
	oldColor := file.At(x, y).(color.RGBA)
	r := oldColor.R << 5 & uint8(0b11100000)
	g := oldColor.G << 3 & uint8(0b00011000)
	b := oldColor.B << 0 & uint8(0b00000111)
	a := r | g | b
	//*list = append(*list, a)
	//mu.Lock()
	(*list)[i] = a
	//mu.Unlock()
}

func GetText(wg *sync.WaitGroup, list []uint8) []string {
	list1 := make([]string, len(list)+1)
	//result := ""
	for c, bin := range list {
		wg.Add(1)
		go func(c1 int, bin1 uint8) {
			defer wg.Done()
			if char, ok := alphabet.ReversedRussianDictionary[encrypt.UintToBinary(bin1, 7)]; ok {
				//result += string(char)
				//mu.Lock()
				list1[c1] = string(char)
				//mu.Unlock()
			} else {
				//result += string(bin)
				//mu.Lock()
				list1[c1] = string(bin1)
				//mu.Unlock()
			}
		}(c, bin)
	}
	wg.Wait()
	return list1
}

func GetPositionBack(wg *sync.WaitGroup, file image.Image, seedOld string) string {
	pol := map[int][]int{
		encrypt.Y1: {encrypt.X1},
	}
	i := uint64(0)
	seed := encrypt.GetSeed(seedOld)
	lenText := GetInfo(file, &pol)
	bounds := file.Bounds()
	width, height := bounds.Max.X, bounds.Max.Y
	list := make([]uint8, lenText)
	rand.Seed(seed)
	for i < lenText {
		y := rand.Intn(height)
		x := rand.Intn(width)
		found := false
		for _, val := range pol[y] {
			if val == x {
				found = true
				break
			}
		}
		if !found {
			pol[y] = append(pol[y], x)
			i++
			wg.Add(1)
			go GetBin(wg, file, x, y, i, &list)
			//list = append(list, a)
		}
	}
	wg.Wait()
	//textList := GetText(wg, list)
	text := strings.Join(GetText(wg, list), "")
	return text
}
