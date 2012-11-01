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

type MetaTag struct {
	String string
    Other []byte
}

func main() {
	flag.Parse()

	log.Printf("Open %s\n", inFile)
	inF, err := os.Open(inFile)
	if err != nil {
		log.Fatal(err)
	}
	defer inF.Close()

	st, _ := inF.Stat()

	header := make([]byte, flv.HEADER_LENGTH)
	_, err = inF.Read(header)
	if (err != nil) {
		log.Fatal(err)
	}

	sig := header[0:3]
	version := header[3:4]
	skip := header[4:5]
	offset := header[5:9]

	log.Printf("InFile %s: Size:%d, SIG:'%s', V:%x, S:%x, O:%x\n", inFile, st.Size(), sig, version, skip, offset)


	next_id := make([]byte, 4)
	_, err = inF.Read(next_id)

	for {
		frame, err := flv.ReadTag(inF)
		if (frame != nil) {
			switch frame.Flavor {
			case flv.KEYFRAME:
				log.Printf("%d\t%d\n", frame.Dts, frame.Position)
			case flv.METADATA:
				buf := bytes.NewReader(frame.Body)
				dec := amf0.NewDecoder(buf)
				objs := []interface{}{}
				for {
					got, err := dec.Decode()
					if err != nil {
						break
					}
					objs = append(objs, got)
				}

				log.Printf("%d\t%d %v\n", frame.Dts, frame.Position, objs[1])

				ea := objs[1].(*amf0.EcmaArrayType)
				for k, v := range (*ea) {
					log.Printf("%v = %v\n", k, v)
				}
				keyframes := (*ea)["keyframes"].(*amf0.ObjectType)

				times := (*keyframes)["times"]
				filepositions := (*keyframes)["filepositions"]

				log.Printf("%v %v\n", times, filepositions)
				return
			}
		}
		if err != nil {
			break
		}
	}

}


