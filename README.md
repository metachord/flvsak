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
     [{<<"stereo">>,true},
      {<<"width">>,960.0},
      {<<"videosize">>,606033659.0},
      {<<"metadatacreator">>,<<"inlet media FLVTool2 v1.0.6 - http://www.inlet-media.de/flvtool2">>},
      {<<"hasMetadata">>,true},
      {<<"audiosamplesize">>,16.0},
      {<<"audiosize">>,62241144.0},
      {<<"hasKeyframes">>,true},
      {<<"audiosamplerate">>,2.2e4},
      {<<"datasize">>,668319089.0},
      {<<"hasCuePoints">>,false},
      {<<"hasAudio">>,true},
      {<<"audiodelay">>,0.0},
      {<<"lastkeyframetimestamp">>,9676.178},
      {<<"framerate">>,14.0},
      {<<"keyframes">>,
       {object,[{times,[0.0,4.0, ...]},
                {filepositions,[44269.0,314034.0, ...]}]}},
      {<<"cuePoints">>,[]},
      {<<"height">>,720.0},
      {<<"audiocodecid">>,2.0},
      {<<"lasttimestamp">>,9677.778},
      {<<"metadatadate">>,{date,1351231793987.901}},
      {<<"filesize">>,670381666.0},
      {<<"canSeekToEnd">>,false},
      {<<"videodatarate">>,499.6493558748713},
      {<<"audiodatarate">>,48.08197170879514},
      {<<"duration">>,9677.845},
      {<<"hasVideo">>,true},
      {<<"videocodecid">>,4.0}]]
```
