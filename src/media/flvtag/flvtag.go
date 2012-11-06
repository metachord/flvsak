package main

import (
	"flag"
	"os"
	"log"
	"bytes"
	"time"
	"math"
	"github.com/metachord/flv.go/flv"
	"github.com/metachord/amf.go/amf0"
)

var inFile string

func init() {
	flag.StringVar(&inFile, "in", "", "input file")
}

type kfTimePos struct {
	Dts uint32
	Position int64
}

func main() {
	flag.Parse()

	log.Printf("Open %s\n", inFile)
	inF, err := os.Open(inFile)
	if err != nil {
		log.Fatal(err)
	}
	defer inF.Close()

	fi, err := inF.Stat()
	if err != nil {
		log.Fatal(err)
	}

	filesize := fi.Size()

	frReader := flv.NewReader(inF)

	_, err = frReader.ReadHeader()
	if err != nil {
		log.Fatal(err)
	}

	var lastKeyFrameTs, lastVTs, lastTs uint32
	var width, height uint16
	var audioRate uint32
	var videoFrameSize, audioFrameSize, dataFrameSize, metadataFrameSize uint64 = 0, 0, 0, 0
	var videoSize, audioSize uint64 = 0, 0
	var videoFrames, audioFrames uint32 = 0, 0

	var kfs []kfTimePos

nextFrame:
	for {
		frame, err := frReader.ReadFrame()
		if (frame != nil) {
			switch frame.(type) {
			case flv.VideoFrame:
				tfr := frame.(flv.VideoFrame)
				width, height = tfr.Width, tfr.Height
				//log.Printf("VideoCodec: %d, Width: %d, Height: %d", tfr.CodecId, tfr.Width, tfr.Height)
				switch tfr.Flavor {
				case flv.KEYFRAME:
					lastKeyFrameTs = tfr.Dts
					kfs = append(kfs, kfTimePos{Dts: tfr.Dts, Position: tfr.Position})
				default:
					videoFrames++
				}
				lastVTs = tfr.Dts
				lastTs = tfr.Dts
				videoFrameSize += uint64(tfr.PrevTagSize)
				videoSize += uint64(len(tfr.Body))
			case flv.AudioFrame:
				tfr := frame.(flv.AudioFrame)
				//log.Printf("AudioCodec: %d, Rate: %d, BitSize: %d, Channels: %d", tfr.CodecId, tfr.Rate, tfr.BitSize, tfr.Channels)
				lastTs = tfr.Dts
				audioRate = tfr.Rate
				audioFrameSize += uint64(tfr.PrevTagSize)
				audioSize += uint64(len(tfr.Body))
				audioFrames++
			case flv.MetaFrame:
				tfr := frame.(flv.MetaFrame)
				buf := bytes.NewReader(tfr.Body)
				dec := amf0.NewDecoder(buf)

				evName, err := dec.Decode()
				if err != nil {
					break nextFrame
				}
				switch evName {
				case amf0.StringType("onMetaData"):

					md, err := dec.Decode()
					if err != nil {
						break nextFrame
					}

					log.Printf("%d\t%d %v\n", tfr.Dts, tfr.Position, md)

					ea := md.(*amf0.EcmaArrayType)
					for k, v := range (*ea) {
						log.Printf("%v = %v\n", k, v)
					}
					keyframes := (*ea)["keyframes"].(*amf0.ObjectType)

					times := (*keyframes)["times"]
					filepositions := (*keyframes)["filepositions"]

					log.Printf("%v %v\n", times, filepositions)
				default:
					log.Printf("Unknown event: %s\n", evName)
				}
				lastTs = tfr.Dts
				metadataFrameSize += uint64(tfr.PrevTagSize)
			}
		}
		if err != nil {
			break
		}
	}
	//log.Printf("KFS: %v", kfs)
	lastKeyFrameTsF := float32(lastKeyFrameTs)/1000
	lastVTsF := float32(lastVTs)/1000
	duration := float32(lastTs)/1000
	dataFrameSize = videoFrameSize + audioFrameSize + metadataFrameSize

	now := time.Now()
	metadatadate := float32(now.Unix() * 1000) + (float32(now.Nanosecond()) / 1000000)

	videoDataRate := (float32(videoSize) / float32(duration))*8/1000
	audioDataRate := (float32(audioSize) / float32(duration))*8/1000

	frameRate := uint8(math.Floor(float64(videoFrames) / float64(duration)))

	log.Printf("FileSize: %d, LastKeyFrameTS: %f, LastTS: %f, Width: %d, Height: %d, VideoSize: %d, AudioSize: %d, MetaDataSize: %d, DataSize: %d, Duration: %f, MetadataDate: %f, VideoDataRate: %f, AudioDataRate: %f, FrameRate: %d, AudioRate: %d", filesize, lastKeyFrameTsF, lastVTsF, width, height, videoFrameSize, audioFrameSize, metadataFrameSize, dataFrameSize, duration, metadatadate, videoDataRate, audioDataRate, frameRate, audioRate)
}

