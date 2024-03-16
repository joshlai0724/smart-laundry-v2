package infoutil

import (
	logutil "edge/util/log"
	"io"
	"os"
)

type Info struct {
	EdgeVersion string
}

func readEdgeVersion(path string) (edgeVersion string, err error) {
	file, err := os.Open(path)
	if err != nil {
		logutil.GetLogger().Warnf("read EDGE_VERSION error, err=%s, path=%s", err, path)
		return
	}
	byteValue, err := io.ReadAll(file)
	if err != nil {
		logutil.GetLogger().Warnf("read EDGE_VERSION error, err=%s, path=%s", err, path)
		return
	}
	return string(byteValue), nil
}

func ReadInfo(edgeVersionPath string) Info {
	edgeVersion, _ := readEdgeVersion(edgeVersionPath)
	return Info{
		EdgeVersion: edgeVersion,
	}
}
