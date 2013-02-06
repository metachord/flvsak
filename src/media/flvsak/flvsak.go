package main

import (
	"bytes"
	"flag"
	"fmt"
	"github.com/metachord/amf.go/amf0"
	"github.com/metachord/flv.go/flv"
	"log"
	"math"
	"os"
	"sort"
	"strconv"
	"strings"
	"time"
)

var inFile string
var outFile string

var printInfo bool

var flvDump bool
var minDts, maxDts int

// comma separated keys
type csKeys []string

var printInfoKeys csKeys

var isConcat bool
var inFiles csKeys

var verbose bool

var updateKeyframes bool

var splitContent bool

// comma separated, map tag type to string
type csTTS map[flv.TagType]string

var outcFiles csTTS

// comma separated, map tag type to int
type csTTI map[flv.TagType]int

var streams csTTI

var fixDts bool

var scaleDts float64 = 1.0

var compensateDts bool

func (i *csKeys) String() string {
	return fmt.Sprint(*i)
}

func (i *csKeys) Set(value string) error {
	for _, mk := range strings.Split(value, ",") {
		*i = append(*i, mk)
	}
	return nil
}

func (i *csTTS) String() string {
	out := make([]string, 0)
	for k, v := range *i {
		out = append(out, fmt.Sprintf("%s:%s", k, v))
	}
	return strings.Join(out, ",")
}

func (i *csTTS) Set(value string) error {
	for _, mk := range strings.Split(value, ",") {

		ts := strings.Split(mk, ":")
		switch ts[0] {
		case "video":
			(*i)[flv.TAG_TYPE_VIDEO] = ts[1]
		case "audio":
			(*i)[flv.TAG_TYPE_AUDIO] = ts[1]
		case "meta":
			(*i)[flv.TAG_TYPE_META] = ts[1]
		default:
			log.Fatalf("Bad content type: %s", ts[0])
		}
	}
	return nil
}

func (i *csTTI) String() string {
	out := make([]string, 0)
	for k, v := range *i {
		var app string
		if v == -1 {
			app = "all"
		} else {
			app = strconv.Itoa(v)
		}
		out = append(out, fmt.Sprintf("%s:%s", k, app))
	}
	return strings.Join(out, ",")
}

func (i *csTTI) Set(value string) error {
	for _, mk := range strings.Split(value, ",") {

		ts := strings.Split(mk, ":")
		switch ts[0] {
		case "video":
			(*i)[flv.TAG_TYPE_VIDEO], _ = strconv.Atoi(ts[1])
		case "audio":
			(*i)[flv.TAG_TYPE_AUDIO], _ = strconv.Atoi(ts[1])
		case "meta":
			(*i)[flv.TAG_TYPE_META], _ = strconv.Atoi(ts[1])
		default:
			log.Fatalf("Bad content type: %s", ts[0])
		}
	}
	return nil
}

func init() {

	outcFiles = make(csTTS)
	outcFiles[flv.TAG_TYPE_VIDEO] = ""
	outcFiles[flv.TAG_TYPE_AUDIO] = ""
	outcFiles[flv.TAG_TYPE_META] = ""

	streams = make(csTTI)
	streams[flv.TAG_TYPE_VIDEO] = -1
	streams[flv.TAG_TYPE_AUDIO] = -1
	streams[flv.TAG_TYPE_META] = -1

	flag.StringVar(&inFile, "in", "", "input file")
	flag.StringVar(&outFile, "out", "", "output file")

	flag.BoolVar(&printInfo, "info", false, "print file info")
	flag.BoolVar(&flvDump, "dump", false, "dump frames")
	flag.IntVar(&minDts, "min-dts", -1, "dump from dts")
	flag.IntVar(&maxDts, "max-dts", -1, "dump to dts")
	flag.Var(&printInfoKeys, "info-keys", "print info from metadata for keys (comma separated)")
	flag.BoolVar(&verbose, "verbose", false, "be verbose")

	flag.BoolVar(&updateKeyframes, "update-keyframes", false, "update keyframes positions in metatag")

	flag.BoolVar(&splitContent, "split-content", false, "split content to different files")

	flag.Var(&outcFiles, "outc", "output frames of declared type to destination")

	flag.BoolVar(&isConcat, "concat", false, "concat files with the same codec")
	flag.Var(&inFiles, "ins", "input files")

	flag.Var(&streams, "streams", "store stream of declared type specified this id (default all)")

	flag.Float64Var(&scaleDts, "scale-dts", 1.0, "scale dts")

	flag.BoolVar(&fixDts, "fix-dts", false, "fix non monotonically dts")
	flag.BoolVar(&compensateDts, "compensate-dts", false, "compensate dts for removed streams")
}

func usage() {
	msg := []string{
		"usage: %s -in in_file.flv",
		" [-update-keyframes -out out_file.flv]",
		" [-info] [-info-keys key1,key2,key3]",
		" [-dump [-min-dts INT] [-max-dts INT]]",
		" [-verbose]",
		" [-fix-dts]",
		" [-split-content [-out-video out_video.flv] [-out-audio out_audio.flv] [-out-meta out_meta.flv]]",
		" [[-stream-video INT] [-stream-audio INT] [-stream-meta INT] [-compensate-dts]]",
		"\n",
	}
	fmt.Fprintf(os.Stderr, strings.Join(msg, "\n"), os.Args[0])
	flag.PrintDefaults()
	os.Exit(2)
}

type kfTimePos struct {
	Dts      uint32
	Position int64
}

func main() {
	flag.Usage = usage
	flag.Parse()

	if isConcat {
		concatFiles()
		return
	}

	if inFile == "" {
		log.Fatal("No input file")
	}

	inF, err := os.Open(inFile)
	if err != nil {
		log.Fatal(err)
	}
	defer inF.Close()

	frReader, header, err := openFrameReader(inF)
	if err != nil {
		log.Fatal(err)
	}

	if printInfo {
		printMetaData(frReader, printInfoKeys)
		return
	} else if flvDump {
		createMetaKeyframes(frReader)
	} else if updateKeyframes {
		if outFile == "" {
			log.Fatal("No output file")
		}

		outF, err := os.Create(outFile)
		if err != nil {
			log.Fatal(err)
		}
		defer outF.Close()

		frWriter := flv.NewWriter(outF)
		frWriter.WriteHeader(header)

		inStart := writeMetaKeyframes(frReader, frWriter)
		inF.Seek(inStart, os.SEEK_SET)

		frW := make(map[flv.TagType]*flv.FlvWriter)
		frW[flv.TAG_TYPE_VIDEO] = frWriter
		frW[flv.TAG_TYPE_AUDIO] = frWriter
		frW[flv.TAG_TYPE_META] = frWriter

		writeFrames(frReader, frW, 0)
	} else if splitContent {
		if outcFiles[flv.TAG_TYPE_VIDEO] == "" && outcFiles[flv.TAG_TYPE_AUDIO] == "" && outcFiles[flv.TAG_TYPE_META] == "" {
			log.Fatal("No any split output file")
		}

		type splitWriter struct {
			FileName string
			Writer   *flv.FlvWriter
		}

		frFW := make(map[flv.TagType]*splitWriter)
		frFW[flv.TAG_TYPE_VIDEO] = &splitWriter{FileName: outcFiles[flv.TAG_TYPE_VIDEO], Writer: nil}
		frFW[flv.TAG_TYPE_AUDIO] = &splitWriter{FileName: outcFiles[flv.TAG_TYPE_AUDIO], Writer: nil}
		frFW[flv.TAG_TYPE_META] = &splitWriter{FileName: outcFiles[flv.TAG_TYPE_META], Writer: nil}

		frW := make(map[flv.TagType]*flv.FlvWriter)

		for k, _ := range frFW {
			var of string
			switch k {
			case flv.TAG_TYPE_VIDEO:
				of = outcFiles[flv.TAG_TYPE_VIDEO]
			case flv.TAG_TYPE_AUDIO:
				of = outcFiles[flv.TAG_TYPE_AUDIO]
			case flv.TAG_TYPE_META:
				of = outcFiles[flv.TAG_TYPE_META]
			}

			var pW *flv.FlvWriter = nil
			for wk, wv := range frFW {
				if wv.FileName == of {
					if wv.Writer != nil {
						log.Printf("Write %s to existing %s file %s", k, wk, of)
						pW = wv.Writer
						break
					}
				}
			}

			if pW != nil {
				frW[k] = pW
			} else {
				outF, err := os.Create(of)
				if err != nil {
					log.Fatal(err)
				}
				log.Printf("Write %s to %s", k, of)
				frFW[k].Writer = flv.NewWriter(outF)
				frFW[k].Writer.WriteHeader(header)
				frW[k] = frFW[k].Writer
			}

		}
		for _, v := range frW {
			defer v.OutFile.Close()
		}
		writeFrames(frReader, frW, 0)
	}
}

func openFrameReader(inF *os.File) (frReader *flv.FlvReader, header *flv.Header, err error) {
	frReader = flv.NewReader(inF)
	header, err = frReader.ReadHeader()
	return
}

func concatFiles() {
	log.Printf("Concat files: %#v", inFiles)
	if outFile == "" {
		log.Fatal("No output file")
	}
	outF, err := os.Create(outFile)
	if err != nil {
		log.Fatal(err)
	}
	defer outF.Close()
	frW := flv.NewWriter(outF)
	frWout := make(map[flv.TagType]*flv.FlvWriter)
	frWout[flv.TAG_TYPE_VIDEO] = frW
	frWout[flv.TAG_TYPE_AUDIO] = frW
	frWout[flv.TAG_TYPE_META] = frW

	wh := true // write header to output after read of first file
	offset := 0
	for _, fn := range inFiles {
		inF, err := os.Open(fn)
		if err != nil {
			log.Fatal(err)
		}
		defer inF.Close()
		frReader, header, err := openFrameReader(inF)
		if err != nil {
			log.Fatal(err)
		}
		if wh {
			frW.WriteHeader(header)
			wh = false
		}

		offset = writeFrames(frReader, frWout, offset)
	}
}

func warnTs(lastTs, stream, currTs uint32) {
	log.Printf("WARN: non monotonically increasing dts in stream %d: %d > %d", stream, lastTs, currTs)
}

func writeFrames(frReader *flv.FlvReader, frW map[flv.TagType]*flv.FlvWriter, offset int) (outOffset int) {
	lastTs := make(map[flv.TagType]map[uint32]uint32)
	lastTsDiff := make(map[flv.TagType]map[uint32]uint32)
	shiftTs := make(map[flv.TagType]map[uint32]uint32)
	for _, c := range []flv.TagType{flv.TAG_TYPE_VIDEO, flv.TAG_TYPE_AUDIO, flv.TAG_TYPE_META} {
		lastTs[c] = make(map[uint32]uint32)
		lastTsDiff[c] = make(map[uint32]uint32)
		shiftTs[c] = make(map[uint32]uint32)
	}

	updateDts := func(cframe flv.Frame) (newDts uint32) {
		c := cframe.GetType()
		s := cframe.GetStream()
		d := cframe.GetDts()
		if lastTs[c][s] > d {
			warnTs(lastTs[c][s], s, d)
			if fixDts {
				newDts := lastTs[c][s] + lastTsDiff[c][s]
				shiftTs[c][s] = newDts - d
				d += shiftTs[c][s]
			}
		}
		d = uint32(int(float64(d)*scaleDts) + offset)
		lastTsDiff[c][s] = d - lastTs[c][s]
		lastTs[c][s] = d
		return d
	}

	var lastInTs uint32 = 0
	var compensateTs uint32 = 0
	for {
		rframe, err := frReader.ReadFrame()
		if err != nil {
			log.Fatal(err)
		}
		if rframe != nil {
			lastInTs = rframe.GetDts()
			if streams[rframe.GetType()] != -1 && rframe.GetStream() != uint32(streams[rframe.GetType()]) {
				if compensateDts {
					compensateTs += (rframe.GetDts() - lastInTs)
				}
				lastInTs = rframe.GetDts()
				continue
			}
			lastInTs = rframe.GetDts()
			if rframe.GetStream() == 0 {
				outOffset = int(lastInTs)
			}
			newDts := updateDts(rframe) - compensateTs
			rframe.SetDts(newDts)
			err = frW[rframe.GetType()].WriteFrame(rframe)

			if err != nil {
				log.Fatal(err)
			}
		} else {
			break
		}
	}
	return
}

func printMetaData(frReader *flv.FlvReader, mk csKeys) {
	_, metaMapP := createMetaKeyframes(frReader)
	metaMap := *metaMapP
	var keys = make(sort.StringSlice, len(metaMap))
	var i int
	for k, _ := range metaMap {
		keys[i] = string(k)
		i++
	}
	sort.Sort(&keys)

	if len(mk) == 0 {
		for i := range keys {
			fmt.Printf("%s: %v\n", keys[i], metaMap[amf0.StringType(keys[i])])
		}
	} else {
		for i := range mk {
			if v, ok := metaMap[amf0.StringType(mk[i])]; ok {
				switch v := v.(type) {
				case *amf0.ObjectType:
					for obk, obv := range *v {
						fmt.Printf("%s[%s]: %v\n", mk[i], obk, obv)
					}
				default:
					fmt.Printf("%s: %v\n", mk[i], v)
				}
			}
		}
	}
}

func writeMetaKeyframes(frReader *flv.FlvReader, frWriter *flv.FlvWriter) (inStart int64) {
	inStart, metaMap := createMetaKeyframes(frReader)

	newBuf := new(bytes.Buffer)
	newEnc := amf0.NewEncoder(newBuf)

	err := newEnc.Encode(amf0.StringType("onMetaData"))
	if err != nil {
		log.Fatalf("%s", err)
	}

	err = newEnc.Encode(metaMap)
	if err != nil {
		log.Fatalf("%s", err)
	}

	cFrame := flv.CFrame{
		Stream: 0,
		Dts:    0,
		Type:   flv.TAG_TYPE_META,
		Flavor: flv.METADATA,
		Body:   newBuf.Bytes(),
	}
	newMdFrame := flv.MetaFrame{
		CFrame: cFrame,
	}

	frWriter.WriteFrame(newMdFrame)
	return inStart
}

func frameDump(fr flv.Frame) {
	if flvDump {
		minValid := (minDts != -1 && fr.GetDts() > uint32(minDts)) || minDts == -1
		maxValid := (maxDts != -1 && fr.GetDts() < uint32(maxDts)) || maxDts == -1
		if minValid && maxValid {
			fmt.Printf("%s\n", fr)
		}
	}
}

func createMetaKeyframes(frReader *flv.FlvReader) (inStart int64, metaMapP *amf0.EcmaArrayType) {

	fi, err := frReader.InFile.Stat()
	if err != nil {
		log.Fatal(err)
	}

	filesize := fi.Size()

	frameSize := map[flv.TagType]uint64{flv.TAG_TYPE_VIDEO: 0, flv.TAG_TYPE_AUDIO: 0, flv.TAG_TYPE_META: 0}
	size := map[flv.TagType]uint64{flv.TAG_TYPE_VIDEO: 0, flv.TAG_TYPE_AUDIO: 0, flv.TAG_TYPE_META: 0}
	has := map[flv.TagType]bool{flv.TAG_TYPE_VIDEO: false, flv.TAG_TYPE_AUDIO: false, flv.TAG_TYPE_META: false}

	var lastKeyFrameTs, lastVTs, lastTs uint32
	var width, height uint16
	var audioRate uint32
	var dataFrameSize uint64 = 0
	var videoFrames, audioFrames uint32 = 0, 0
	var stereo bool = false
	var videoCodec, audioCodec uint8 = 0, 0
	var audioSampleSize uint32 = 0
	var hasKeyframes bool = false

	var oldOnMetaDataSize int64 = 0

	var kfs []kfTimePos

nextFrame:
	for {
		frame, err := frReader.ReadFrame()
		if frame != nil {
			switch tfr := frame.(type) {
			case flv.VideoFrame:
				if (width == 0) || (height == 0) {
					width, height = tfr.Width, tfr.Height
					//log.Printf("VideoCodec: %d, Width: %d, Height: %d", tfr.CodecId, tfr.Width, tfr.Height)
				}
				switch tfr.Flavor {
				case flv.KEYFRAME:
					lastKeyFrameTs = tfr.Dts
					hasKeyframes = true
					kfs = append(kfs, kfTimePos{Dts: tfr.Dts, Position: tfr.Position})
				default:
					videoFrames++
				}
				lastVTs = tfr.Dts
				videoCodec = uint8(tfr.CodecId)
			case flv.AudioFrame:
				audioRate = tfr.Rate
				if tfr.Channels == flv.AUDIO_TYPE_STEREO {
					stereo = true
				}
				switch tfr.BitSize {
				case flv.AUDIO_SIZE_8BIT:
					audioSampleSize = 8
				case flv.AUDIO_SIZE_16BIT:
					audioSampleSize = 16
				}
				audioCodec = uint8(tfr.CodecId)
				audioFrames++
			case flv.MetaFrame:
				buf := bytes.NewReader(tfr.Body)
				dec := amf0.NewDecoder(buf)

				evName, err := dec.Decode()
				if err != nil {
					break nextFrame
				}
				switch evName {
				case amf0.StringType("onMetaData"):
					oldOnMetaDataSize = int64(tfr.PrevTagSize)
					md, err := dec.Decode()
					if err != nil {
						break nextFrame
					}

					var ea map[amf0.StringType]interface{}
					switch md := md.(type) {
					case *amf0.EcmaArrayType:
						ea = *md
					case *amf0.ObjectType:
						ea = *md
					}
					if verbose {
						log.Printf("Old onMetaData")
						for k, v := range ea {
							log.Printf("%v = %v\n", k, v)
						}
					}
					if width == 0 {
						if v, ok := ((ea)["width"]); ok {
							width = uint16(v.(amf0.NumberType))
						}
					}
					if height == 0 {
						if v, ok := ((ea)["height"]); ok {
							height = uint16(v.(amf0.NumberType))
						}
					}
				default:
					log.Printf("Unknown event: %s\n", evName)
				}
			}
			frameSize[frame.GetType()] += uint64(frame.GetPrevTagSize())
			size[frame.GetType()] += uint64(len(*frame.GetBody()))
			has[frame.GetType()] = true
			lastTs = frame.GetDts()
			frameDump(frame)
		} else {
			break
		}
		if err != nil {
			break
		}
	}
	if flvDump && !printInfo {
		return 0, nil
	}

	//log.Printf("KFS: %v", kfs)
	lastKeyFrameTsF := float32(lastKeyFrameTs) / 1000
	lastVTsF := float32(lastVTs) / 1000
	duration := float32(lastTs) / 1000
	dataFrameSize = frameSize[flv.TAG_TYPE_VIDEO] + frameSize[flv.TAG_TYPE_AUDIO] + frameSize[flv.TAG_TYPE_META]

	now := time.Now()
	metadatadate := float64(now.Unix()*1000) + (float64(now.Nanosecond()) / 1000000)

	videoDataRate := (float32(size[flv.TAG_TYPE_VIDEO]) / float32(duration)) * 8 / 1000
	audioDataRate := (float32(size[flv.TAG_TYPE_AUDIO]) / float32(duration)) * 8 / 1000

	frameRate := uint8(math.Floor(float64(videoFrames) / float64(duration)))

	//log.Printf("oldOnMetaDataSize: %d, FileSize: %d, LastKeyFrameTS: %f, LastTS: %f, Width: %d, Height: %d, VideoSize: %d, AudioSize: %d, MetaDataSize: %d, DataSize: %d, Duration: %f, MetadataDate: %f, VideoDataRate: %f, AudioDataRate: %f, FrameRate: %d, AudioRate: %d", oldOnMetaDataSize, filesize, lastKeyFrameTsF, lastVTsF, width, height, videoFrameSize, audioFrameSize, metadataFrameSize, dataFrameSize, duration, metadatadate, videoDataRate, audioDataRate, frameRate, audioRate)

	kfTimes := make(amf0.StrictArrayType, 0)
	kfPositions := make(amf0.StrictArrayType, 0)

	for i := range kfs {
		kfTimes = append(kfTimes, amf0.NumberType((float64(kfs[i].Dts) / 1000)))
		kfPositions = append(kfTimes, amf0.NumberType(kfs[i].Position))
	}

	keyFrames := amf0.ObjectType{
		"times":         &kfTimes,
		"filepositions": &kfPositions,
	}

	has[flv.TAG_TYPE_META] = true

	metaMap := amf0.EcmaArrayType{
		"metadatacreator": amf0.StringType("FlvSAK https://github.com/metachord/flvsak"),
		"metadatadate":    amf0.DateType{TimeZone: 0, Date: metadatadate},

		"keyframes": &keyFrames,

		"hasVideo":     amf0.BooleanType(has[flv.TAG_TYPE_VIDEO]),
		"hasAudio":     amf0.BooleanType(has[flv.TAG_TYPE_AUDIO]),
		"hasMetadata":  amf0.BooleanType(has[flv.TAG_TYPE_META]),
		"hasKeyframes": amf0.BooleanType(hasKeyframes),
		"hasCuePoints": amf0.BooleanType(false),

		"videocodecid":  amf0.NumberType(videoCodec),
		"width":         amf0.NumberType(width),
		"height":        amf0.NumberType(height),
		"videosize":     amf0.NumberType(frameSize[flv.TAG_TYPE_VIDEO]),
		"framerate":     amf0.NumberType(frameRate),
		"videodatarate": amf0.NumberType(videoDataRate),

		"audiocodecid":    amf0.NumberType(audioCodec),
		"stereo":          amf0.BooleanType(stereo),
		"audiosamplesize": amf0.NumberType(audioSampleSize),
		"audiodelay":      amf0.NumberType(0),
		"audiodatarate":   amf0.NumberType(audioDataRate),
		"audiosize":       amf0.NumberType(frameSize[flv.TAG_TYPE_AUDIO]),
		"audiosamplerate": amf0.NumberType(audioRate),

		"filesize":              amf0.NumberType(filesize),
		"datasize":              amf0.NumberType(dataFrameSize),
		"lasttimestamp":         amf0.NumberType(lastVTsF),
		"lastkeyframetimestamp": amf0.NumberType(lastKeyFrameTsF),
		"cuePoints":             &amf0.StrictArrayType{},
		"duration":              amf0.NumberType(duration),
		"canSeekToEnd":          amf0.BooleanType(false),
	}

	if verbose {
		log.Printf("New onMetaData")
		for k, v := range metaMap {
			log.Printf("%v = %v\n", k, v)
		}
	}

	buf := new(bytes.Buffer)
	enc := amf0.NewEncoder(buf)
	err = enc.Encode(&metaMap)
	if err != nil {
		log.Fatalf("%s", err)
	}

	newOnMetaDataSize := int64(buf.Len()) + int64(flv.TAG_HEADER_LENGTH) + int64(flv.PREV_TAG_SIZE_LENGTH)
	//log.Printf("newOnMetaDataSize: %v", newOnMetaDataSize)
	//log.Printf("oldKeyFrames: %v", &keyFrames)

	newKfPositions := make(amf0.StrictArrayType, 0)

	var dataDiff int64 = newOnMetaDataSize - oldOnMetaDataSize

	for i := range kfs {
		newKfPositions = append(newKfPositions, amf0.NumberType(uint64(kfs[i].Position+dataDiff)))
	}
	keyFrames["filepositions"] = &newKfPositions
	metaMap["filesize"] = amf0.NumberType(int64(metaMap["filesize"].(amf0.NumberType)) + dataDiff)
	metaMap["datasize"] = amf0.NumberType(int64(metaMap["datasize"].(amf0.NumberType)) + dataDiff)

	//log.Printf("newKeyFrames: %v", &keyFrames)

	inStart = kfs[0].Position
	return inStart, &metaMap
}
