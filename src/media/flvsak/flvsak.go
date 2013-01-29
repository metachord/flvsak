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

var videoOutFile string
var audioOutFile string
var metaOutFile string

var streamVideo, streamAudio, streamMeta int

var fixDts bool

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

func init() {

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
	flag.StringVar(&videoOutFile, "out-video", "", "output video file")
	flag.StringVar(&audioOutFile, "out-audio", "", "output audio file")
	flag.StringVar(&metaOutFile, "out-meta", "", "output meta file")

	flag.BoolVar(&isConcat, "concat", false, "concat files with the same codec")
	flag.Var(&inFiles, "ins", "input files")

	flag.IntVar(&streamVideo, "stream-video", -1, "store video stream with this id (default all)")
	flag.IntVar(&streamAudio, "stream-audio", -1, "store audio stream with this id (default all)")
	flag.IntVar(&streamMeta, "stream-meta", -1, "store meta stream with this id (default all)")

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

		frW := make(map[string]*flv.FlvWriter)
		frW["video"] = frWriter
		frW["audio"] = frWriter
		frW["meta"] = frWriter

		writeFrames(frReader, frW, 0)
	} else if splitContent {
		if videoOutFile == "" && audioOutFile == "" && metaOutFile == "" {
			log.Fatal("No any split output file")
		}

		type splitWriter struct {
			FileName string
			Writer   *flv.FlvWriter
		}

		frFW := make(map[string]*splitWriter)
		frFW["video"] = &splitWriter{FileName: videoOutFile, Writer: nil}
		frFW["audio"] = &splitWriter{FileName: audioOutFile, Writer: nil}
		frFW["meta"] = &splitWriter{FileName: metaOutFile, Writer: nil}

		frW := make(map[string]*flv.FlvWriter)

		for k, _ := range frFW {
			var of string
			switch k {
			case "video":
				of = videoOutFile
			case "audio":
				of = audioOutFile
			case "meta":
				of = metaOutFile
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
	frWout := make(map[string]*flv.FlvWriter)
	frWout["audio"] = frW
	frWout["video"] = frW
	frWout["meta"] = frW

	wh := true					// write header to output after read of first file
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

func writeFrames(frReader *flv.FlvReader, frW map[string]*flv.FlvWriter, offset int) (outOffset int) {
	lastTs := make(map[string]map[uint32]uint32)
	lastTsDiff := make(map[string]map[uint32]uint32)
	shiftTs := make(map[string]map[uint32]uint32)
	for _, c := range []string{"video", "audio", "meta"} {
		lastTs[c] = make(map[uint32]uint32)
		lastTsDiff[c] = make(map[uint32]uint32)
		shiftTs[c] = make(map[uint32]uint32)
	}

	updateDts := func(c string, s uint32, d uint32) (newDts uint32) {
		if lastTs[c][s] > d {
			warnTs(lastTs[c][s], s, d)
			if fixDts {
				newDts := lastTs[c][s] + lastTsDiff[c][s]
				shiftTs[c][s] = newDts - d
				d += shiftTs[c][s]
			}
		}
		d = uint32(int(d) + offset)
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
			switch rframe.(type) {
			case flv.VideoFrame:
				f := rframe.(flv.VideoFrame)
				lastInTs = f.Dts
				if streamVideo != -1 && f.Stream != uint32(streamVideo) {
					if compensateDts {
						compensateTs += (f.Dts - lastInTs)
					}
					lastInTs = f.Dts
					continue
				}
				c := "video"
				lastInTs = f.Dts
				if f.Stream == 0 {
					outOffset = int(lastInTs)
				}
				f.Dts = updateDts(c, f.Stream, f.Dts) - compensateTs
				err = frW[c].WriteFrame(f)
			case flv.AudioFrame:
				f := rframe.(flv.AudioFrame)
				lastInTs = f.Dts
				if streamAudio != -1 && f.Stream != uint32(streamAudio) {
					if compensateDts {
						compensateTs += (f.Dts - lastInTs)
					}
					lastInTs = f.Dts
					continue
				}
				c := "audio"
				lastInTs = f.Dts
				if f.Stream == 0 {
					outOffset = int(lastInTs)
				}
				f.Dts = updateDts(c, f.Stream, f.Dts) - compensateTs
				err = frW[c].WriteFrame(f)
			case flv.MetaFrame:
				f := rframe.(flv.MetaFrame)
				lastInTs = f.Dts
				if streamMeta != -1 && f.Stream != uint32(streamMeta) {
					if compensateDts {
						compensateTs += (f.Dts - lastInTs)
					}
					lastInTs = f.Dts
					continue
				}
				c := "meta"
				lastInTs = f.Dts
				if f.Stream == 0 {
					outOffset = int(lastInTs)
				}
				f.Dts = updateDts(c, f.Stream, f.Dts) - compensateTs
				err = frW[c].WriteFrame(f)
			}
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

	var lastKeyFrameTs, lastVTs, lastTs uint32
	var width, height uint16
	var audioRate uint32
	var videoFrameSize, audioFrameSize, dataFrameSize, metadataFrameSize uint64 = 0, 0, 0, 0
	var videoSize, audioSize uint64 = 0, 0
	var videoFrames, audioFrames uint32 = 0, 0
	var stereo bool = false
	var videoCodec, audioCodec uint8 = 0, 0
	var audioSampleSize uint32 = 0
	var hasVideo, hasAudio, hasMetadata, hasKeyframes bool = false, false, false, false

	var oldOnMetaDataSize int64 = 0

	var kfs []kfTimePos

nextFrame:
	for {
		frame, err := frReader.ReadFrame()
		if frame != nil {
			switch frame.(type) {
			case flv.VideoFrame:
				tfr := frame.(flv.VideoFrame)
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
				frameDump(tfr)
				hasVideo = true
				lastVTs = tfr.Dts
				lastTs = tfr.Dts
				videoCodec = uint8(tfr.CodecId)
				videoFrameSize += uint64(tfr.PrevTagSize)
				videoSize += uint64(len(tfr.Body))
			case flv.AudioFrame:
				tfr := frame.(flv.AudioFrame)
				//log.Printf("AudioCodec: %d, Rate: %d, BitSize: %d, Channels: %d", tfr.CodecId, tfr.Rate, tfr.BitSize, tfr.Channels)
				//lastATs = tfr.Dts
				lastTs = tfr.Dts
				audioRate = tfr.Rate
				audioFrameSize += uint64(tfr.PrevTagSize)
				audioSize += uint64(len(tfr.Body))
				if tfr.Channels == flv.AUDIO_TYPE_STEREO {
					stereo = true
				}
				switch tfr.BitSize {
				case flv.AUDIO_SIZE_8BIT:
					audioSampleSize = 8
				case flv.AUDIO_SIZE_16BIT:
					audioSampleSize = 16
				}
				frameDump(tfr)
				hasAudio = true
				audioCodec = uint8(tfr.CodecId)
				audioFrames++
			case flv.MetaFrame:
				tfr := frame.(flv.MetaFrame)
				buf := bytes.NewReader(tfr.Body)
				dec := amf0.NewDecoder(buf)

				hasMetadata = true
				lastTs = tfr.Dts
				metadataFrameSize += uint64(tfr.PrevTagSize)
				frameDump(tfr)
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
	dataFrameSize = videoFrameSize + audioFrameSize + metadataFrameSize

	now := time.Now()
	metadatadate := float64(now.Unix()*1000) + (float64(now.Nanosecond()) / 1000000)

	videoDataRate := (float32(videoSize) / float32(duration)) * 8 / 1000
	audioDataRate := (float32(audioSize) / float32(duration)) * 8 / 1000

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

	hasMetadata = true

	metaMap := amf0.EcmaArrayType{
		"metadatacreator": amf0.StringType("FlvSAK https://github.com/metachord/flvsak"),
		"metadatadate":    amf0.DateType{TimeZone: 0, Date: metadatadate},

		"keyframes": &keyFrames,

		"hasVideo":     amf0.BooleanType(hasVideo),
		"hasAudio":     amf0.BooleanType(hasAudio),
		"hasMetadata":  amf0.BooleanType(hasMetadata),
		"hasKeyframes": amf0.BooleanType(hasKeyframes),
		"hasCuePoints": amf0.BooleanType(false),

		"videocodecid":  amf0.NumberType(videoCodec),
		"width":         amf0.NumberType(width),
		"height":        amf0.NumberType(height),
		"videosize":     amf0.NumberType(videoFrameSize),
		"framerate":     amf0.NumberType(frameRate),
		"videodatarate": amf0.NumberType(videoDataRate),

		"audiocodecid":    amf0.NumberType(audioCodec),
		"stereo":          amf0.BooleanType(stereo),
		"audiosamplesize": amf0.NumberType(audioSampleSize),
		"audiodelay":      amf0.NumberType(0),
		"audiodatarate":   amf0.NumberType(audioDataRate),
		"audiosize":       amf0.NumberType(audioFrameSize),
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
