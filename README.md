# flvtag #

Tool for tagging flv files

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
