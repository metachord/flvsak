# flvtag #

Tool for tagging flv files


## Benchmark ##

```
    $ ls -l in.flv
    -rw-r--r-- 1 root root 398312988 2012-11-06 10:58 in.flv

    $ time flvtag -in in.flv -out out.flv

    real    0m1.803s
    user    0m0.640s
    sys     0m1.170s


    $ time flvtool2 -U in.flv

    real    8m37.501s
    user    8m32.050s
    sys     0m4.900s
```

## Description ##

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
    metadatacreator = Flvtag https://github.com/metachord/flvtag
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
