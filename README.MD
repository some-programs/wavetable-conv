# HELLO

## What is this?

This code is highly specific to converting the 256 samples per waveform
wavetables from https://waveeditonline.com/ to 1024 samples per waveform
wavetables to use with the Ableton Wavetable music synthesizer.

NOTE: Right now the tool has some hard coded values to the specific sample
rates and lenghts of the input files so the source has to be modified to
process other files. This is just a bunch of files that has been lying around
my computer for a while and I don't have immediate plans to add features for
other uses cases.

## Tool Output files

Two output files are generated for each input:

Suffix _m : Each sample is multiplied by 4, no resampling.

Suffix _r: Reasampled using libsoxr's highest quality mode.

Some times the multiplied wavetables sounds more interesting and some times
they are just nice lofi alternatives to the resampled ones.

## Directories

/waveeditonline/ : just a copy of files from https://waveeditonline.com/

/ableton-wavetable/ : output files to use with Ableton Wavetable (make sure to click the RAW button)


