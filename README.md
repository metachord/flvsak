# flvsak — the Swiss Army Knife for FLV files #

Tool for different operations on FLV files. Initially created to replace extremely slow flvtool2 -U.

## Update keyframes ##

### Benchmark ###

Update metadata keyframes records:

```
    $ ls -l in.flv
    -rw-r--r-- 1 root root 398312988 2012-11-06 10:58 in.flv

    $ time flvsak -in in.flv -out out.flv -update-keyframes

    real    0m1.803s
    user    0m0.640s
    sys     0m1.170s


    $ time flvtool2 -U in.flv

    real    8m37.501s
    user    8m32.050s
    sys     0m4.900s
```

### Description ###

Add following values to metadata:

 * audiocodecid
 * audiodatarate
 * audiodelay
 * audiosamplerate
 * audiosamplesize
 * audiosize
 * canSeekToEnd
 * cuePoints
 * duration
 * framerate
 * hasAudio
 * hasCuePoints
 * hasKeyframes
 * hasVideo
 * height
 * lastkeyframetimestamp
 * lasttimestamp
 * stereo
 * videocodecid
 * videodatarate
 * videosize
 * width


Following records are recalculated after building first metadata tag:

 * datasize
 * filesize
 * hasMetadata
 * keyframes

Example of metadata:

```
    [<<"onMetaData">>,
     [
      * {<<"width">>,960.0},
      * {<<"filesize">>,458356236.0},
      * {<<"lasttimestamp">>,6610.245},
      * {<<"lastkeyframetimestamp">>,6609.445},
      * {<<"metadatacreator">>,<<"inlet media FLVTool2 v1.0.6 - http://www.inlet-media.de/flvtool2">>},
      * {<<"hasKeyframes">>,true},
      * {<<"audiodatarate">>,48.082306177758916},
      * {<<"videosize">>,414403791.0},
      * {<<"canSeekToEnd">>,false},
      * {<<"videocodecid">>,4.0},
      * {<<"audiocodecid">>,2.0},
      * {<<"videodatarate">>,500.20897803334066},
      * {<<"height">>,720.0},
      * {<<"datasize">>,456947343.0},
      * {<<"hasAudio">>,true},
      * {<<"framerate">>,14.0},
      * {<<"audiosamplesize">>,16.0},
      * {<<"metadatadate">>,{date,1351681379738.907}},
      * {<<"hasCuePoints">>,false},
      * {<<"cuePoints">>,[]},
      * {<<"duration">>,6610.312},
      * {<<"hasVideo">>,true},
      * {<<"audiosamplerate">>,2.2e4},
      * {<<"stereo">>,true},
      * {<<"keyframes">>,
         {object,[{times,[0.0,4.0, ...]},
                  {filepositions,[30463.0,287756.0, ...]}]}},
      * {<<"audiodelay">>,0.0},
      * {<<"audiosize">>,42513072.0},
      * {<<"hasMetadata">>,true}
      ]]
```


```
    lastkeyframetimestamp = 6609.44482421875
    audiodatarate = 48.081443786621094
    videocodecid = 4
    audiodelay = 0
    width = 960
    metadatacreator = FlvSAK https://github.com/metachord/flvsak
    hasVideo = true
    audiosize = 4.2513072e+07
    canSeekToEnd = false
    duration = 6610.36376953125
    videosize = 4.14403791e+08
    hasKeyframes = true
    filesize = 4.58356236e+08
    audiosamplesize = 16
    lasttimestamp = 6610.2451171875
    metadatadate = {0 1.352184840548653e+12}
    height = 720
    framerate = 14
    keyframes = 0xf840049658
    stereo = true
    cuePoints = &[]
    datasize = 4.56947343e+08
    audiocodecid = 2
    hasMetadata = true
    hasCuePoints = false
    audiosamplerate = 22050
    hasAudio = true
    videodatarate = 500.20001220703125

```

## Print metadata entries ##

All metadata in alphabetically order:

```
    $ flvsak -in in_file.flv -info
    audiocodecid: 2
    audiodatarate: 44.8125
    audiodelay: 0
    audiosamplerate: 22000
    audiosamplesize: 16
    audiosize: 4.2513072e+07
    canSeekToEnd: false
    cuePoints: &[]
    datasize: 4.56947287e+08
    duration: 7092.57080078125
    filesize: 4.58356172e+08
    framerate: 13
    hasAudio: true
    hasCuePoints: false
    hasKeyframes: true
    hasMetadata: true
    hasVideo: true
    height: 720
    keyframes: 0xf840047348
    lastkeyframetimestamp: 7091.72900390625
    lasttimestamp: 7092.533203125
    metadatacreator: FlvSAK https://github.com/metachord/flvsak
    metadatadate: {0 1.354265321490319e+12}
    stereo: true
    videocodecid: 4
    videodatarate: 466.1925964355469
    videosize: 4.14403791e+08
    width: 960
```

Specify list of keys from metadata:

```
    $ flvsak -in in_file.flv -info -info-keys height,width,duration
    height: 720
    width: 960
    duration: 7092.57080078125
```

## Dump frames ##

Dump all frames info to stdout between `min-dts` and `max-dts`:

```
    $ flvsak -in in_file.flv -dump -min-dts 6031657 -max-dts 7092449
```

## Split content to different files ##

The following command will split `in_file.flv` to two files: `out.flv` (contains only audio and video only for stream `0`) and `out-meta.flv` (contains all metadata for all streams). Flag `-fix-dts` will fix non monotonically increasing DTS in input file.

```
    $ flvsak -in in_file.flv -split-content -outc video:out.flv,audio:out.flv,meta:out-meta.flv -fix-dts -streams video:0,audio:0
```

## Crop file by DTS ##

Crop parts of file in specified ranges of DTS. Flag `-crop-wait-keyframe` will crop at nearest keyframe.

```
    $ flvsak -in in_file.flv -out out_crop.flv -crop 1619000..1731000,2753000..2812000 -crop-wait-keyframe
```

If crop range is one number, frame with this DTS will be cropped.

## Broken file recover ##

To recover broken FLV-file, use flag `-recover`. On every broken frame reader will skip byte until valid. If option `-max-frame-size` specified frame with body greater than this value also broken.

```
    $ ./flvsak/flvsak -in in_file.flv -split-content -outc video:out.flv,audio:out.flv,meta:out.flv -recover -max-frame-size 100000
```
