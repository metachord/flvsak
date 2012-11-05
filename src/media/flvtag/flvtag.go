package main

import (
	"flag"
	"os"
	"log"
	"bytes"
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

	frReader := flv.NewReader(inF)

	_, err = frReader.ReadHeader()
	if err != nil {
		log.Fatal(err)
	}

	var kfs []kfTimePos


nextFrame:
	for {
		frame, err := frReader.ReadFrame()
		if (frame != nil) {
			switch frame.(type) {
			case flv.VideoFrame:
				tfr := frame.(flv.VideoFrame)
				log.Printf("VideoCodec: %d, Width: %d, Height: %d", tfr.CodecId, tfr.Width, tfr.Height)
				switch tfr.Flavor {
				case flv.KEYFRAME:
					kfs = append(kfs, kfTimePos{Dts: tfr.Dts, Position: tfr.Position})
				}
			case flv.AudioFrame:
				tfr := frame.(flv.AudioFrame)
				log.Printf("AudioCodec: %d, Rate: %d, BitSize: %d, Channels: %d", tfr.CodecId, tfr.Rate, tfr.BitSize, tfr.Channels)
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
			}
		}
		if err != nil {
			break
		}
	}
	log.Printf("KFS: %v", kfs)
}

