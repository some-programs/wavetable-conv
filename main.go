package main

// This is a quick hack to get exactly one job done, not much care given to
// good engingeering practices.

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"io"
	"io/fs"
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/youpy/go-wav"
	"github.com/zaf/resample"
)

const inRoot = "waveeditonline"
const outRoot = "out"

func main() {
	log.SetFlags(log.Lshortfile)
	os.RemoveAll(outRoot)
	renames()
	process()
}

// process handles the conversion processing
func process() {
	wavs := os.DirFS(inRoot)
	err := fs.WalkDir(wavs, ".", func(path string, d fs.DirEntry, err error) error {
		if d.IsDir() {
			return nil
		}

		in := filepath.Join(inRoot, path)
		out := filepath.Join(outRoot, path)
		_ = os.MkdirAll(filepath.Dir(out), 0700)

		log.SetPrefix(path + " ")
		log.Println("start processing")

		err = multiply(in, suffix(out, "_m"))
		if err != nil {
			return err
		}

		err = resamp(in, suffix(out, "_r"))
		if err != nil {
			return err
		}

		err = resampWhole(in, suffix(out, "_w"))
		if err != nil {
			return err
		}

		return nil
	})
	log.SetPrefix("")
	if err != nil {
		panic(err)
	}

}

// suffix returns a filename with suffix added.
func suffix(name string, suffix string) string {
	ext := filepath.Ext(name)
	prefix := strings.TrimSuffix(name, ext)
	return prefix + suffix + ext
}

func resamp(inName, outName string) error {

	f, err := os.Open(inName)
	if err != nil {
		return err
	}
	defer f.Close()

	reader := wav.NewReader(f)
	format, err := reader.Format()
	if err != nil {
		return err
	}
	// log.Println("wav-in")
	// log.Println(spew.Sdump(format))

	var samples []wav.Sample
	for {

		ss, err := reader.ReadSamples()
		if err != nil {
			if err == io.EOF {
				break
			}
			return err
		}
		samples = append(samples, ss...)

	}

	// log.Println("samples len", len(samples))

	if len(samples) != 16384 {
		log.Fatalln("unexpected wave file length", len(samples))
	}

	var waveforms [][]byte
	enc := binary.LittleEndian

	{
		for wfi := 0; wfi < len(samples); wfi += 256 {
			var waveform []byte
			for si := 0; si < 256; si++ {
				s := samples[wfi+si]
				bs := make([]byte, 8)
				enc.PutUint64(bs, uint64(s.Values[0]))
				bs = bs[:(format.BitsPerSample / 8)]
				waveform = append(waveform, bs...)

			}
			waveforms = append(waveforms, waveform)
		}

		// log.Printf("%s %s", outName, strings.Join(strings.Split(fmt.Sprintf("%+x", waveforms), " "), "\n\n"))
	}

	var resampled [][]byte

	for _, wf := range waveforms {
		// log.Println(i, len(wf))

		var b bytes.Buffer

		res, err := resample.New(
			&b,
			float64(format.SampleRate),
			float64(format.SampleRate*4),
			1,
			resample.I16,
			resample.VeryHighQ,
		)
		if err != nil {
			return err
		}

		_, err = res.Write(wf)
		if err != nil {
			return err
		}
		res.Close()
		// log.Println("len", b.Len())
		resampled = append(resampled, b.Bytes())
	}

	{
		var count, wfcount int
		for idx, w := range resampled {
			count += len(w)
			wfcount += len(waveforms[idx])
		}
		log.Printf("samples %v | waveforms %v (%v bytes) | resampled %v (%v bytes)",
			len(samples),
			len(waveforms), wfcount,
			len(resampled), count,
		)
	}

	outf, err := os.OpenFile(outName, os.O_CREATE|os.O_TRUNC|os.O_WRONLY, 0600)
	if err != nil {
		return err
	}
	defer outf.Close()

	writer := wav.NewWriter(
		outf,
		uint32(len(samples)*4),
		// numSamples uint32,
		format.NumChannels,
		format.SampleRate,
		format.BitsPerSample,
	)

	for _, wf := range resampled {
		_, err := writer.Write(wf)
		if err != nil {
			return err
		}
	}
	// log.Println("wav-out")
	// spew.Dump(writer.Format)

	return nil

}

func resampWhole(inName, outName string) error {
	f, err := os.Open(inName)
	if err != nil {
		return err
	}
	defer f.Close()

	reader := wav.NewReader(f)
	format, err := reader.Format()
	if err != nil {
		return err
	}

	// log.Println("wav-in")
	// log.Println(spew.Sdump(format))

	var samples []wav.Sample
	for {

		ss, err := reader.ReadSamples()
		if err != nil {
			if err == io.EOF {
				break
			}
			return err
		}
		samples = append(samples, ss...)

	}

	// log.Println("samples len", len(samples))

	if len(samples) != 16384 {
		log.Fatalln("unexpected wave file length", len(samples))
	}

	var sampleBytes []byte
	enc := binary.LittleEndian
	for wfi := 0; wfi < len(samples); wfi++ {
		s := samples[wfi]
		bs := make([]byte, 8)
		enc.PutUint64(bs, uint64(s.Values[0]))
		bs = bs[:(format.BitsPerSample / 8)]
		sampleBytes = append(sampleBytes, bs...)
	}

	var resampled []byte
	{
		var b bytes.Buffer
		res, err := resample.New(
			&b,
			float64(format.SampleRate),
			float64(format.SampleRate*4),
			1,
			resample.I16,
			resample.VeryHighQ,
		)
		if err != nil {
			return err
		}

		{
			_, err := res.Write(sampleBytes)
			if err != nil {
				return err
			}
		}

		res.Close()
		resampled = b.Bytes()
	}

	outf, err := os.OpenFile(outName, os.O_CREATE|os.O_TRUNC|os.O_WRONLY, 0600)
	if err != nil {
		return err
	}
	defer outf.Close()

	writer := wav.NewWriter(
		outf,
		uint32(len(samples)*4),
		// numSamples uint32,
		format.NumChannels,
		format.SampleRate,
		format.BitsPerSample,
	)

	{
		_, err := writer.Write(resampled)
		if err != nil {
			return err
		}
	}

	// log.Println("wav-out")
	// spew.Dump(writer.Format)

	return nil

}

// Multiply simply quadruples all samples to get from 256 to 1024 samples per waveform.
// This both sounds much better for some waveforms and/or adds an lofi version
// of other waveforms that can be sonically interesting.
func multiply(inName, outName string) error {
	f, err := os.Open(inName)
	if err != nil {
		return err
	}
	defer f.Close()

	reader := wav.NewReader(f)
	format, err := reader.Format()
	if err != nil {
		return err
	}

	var samples []wav.Sample
	for {
		ss, err := reader.ReadSamples()
		if err != nil {
			if err == io.EOF {
				break
			}
			return err
		}
		samples = append(samples, ss...)
	}

	outf, err := os.OpenFile(outName, os.O_CREATE|os.O_TRUNC|os.O_WRONLY, 0600)
	if err != nil {
		return err
	}
	defer outf.Close()

	writer := wav.NewWriter(
		outf,
		uint32(len(samples)*4),
		// numSamples uint32,
		format.NumChannels,
		format.SampleRate,
		format.BitsPerSample,
	)
	for _, s := range samples {
		ss := []wav.Sample{s, s, s, s}
		if err := writer.WriteSamples(ss); err != nil {
			return err
		}
	}

	return nil
}

// renames and lightly organize files into directories from a
// waveeditonline.com all WAV at once zip file.
func renames() {
	{
		de, err := os.ReadDir(inRoot)
		if err != nil {
			panic(err)
		}
		for _, e := range de {
			if e.IsDir() {
				continue
			}
			name := e.Name()
			low := strings.ToLower(name)
			if name != low {
				fmt.Println(name, "->", low)
				err := os.Rename(filepath.Join(inRoot, name), filepath.Join(inRoot, low))
				if err != nil {
					panic(err)
				}
			}
		}
	}
	{
		prefixes := map[string]string{
			"grav-":  "grav",
			"ppg_":   "ppg",
			"sohler": "sohler",
			"synlp":  "synlp",
			"synlpg": "synlp",
			"tidyb":  "tidyb",
		}
		de, err := os.ReadDir(inRoot)
		if err != nil {
			panic(err)
		}
		for _, e := range de {
			if e.IsDir() {
				continue
			}
			name := e.Name()
			subdir := "misc"
		pfs:
			for k, v := range prefixes {
				if strings.HasPrefix(name, k) {
					subdir = v
					break pfs
				}
			}
			os.MkdirAll(filepath.Join(inRoot, subdir), 0700)

			err := os.Rename(filepath.Join(inRoot, name), filepath.Join(inRoot, subdir, name))
			if err != nil {
				panic(err)
			}
		}

	}

}
