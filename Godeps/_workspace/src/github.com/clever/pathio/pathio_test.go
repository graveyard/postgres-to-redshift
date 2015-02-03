package pathio

import (
	"bufio"
	"io/ioutil"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestParseS3Path(t *testing.T) {
	bucketName, s3path, err := parseS3Path("s3://clever-files/directory/path")
	assert.Nil(t, err)
	assert.Equal(t, bucketName, "clever-files")
	assert.Equal(t, s3path, "directory/path")

	bucketName, s3path, err = parseS3Path("s3://clever-files/directory")
	assert.Nil(t, err)
	assert.Equal(t, bucketName, "clever-files")
	assert.Equal(t, s3path, "directory")
}

func TestParseInvalidS3Path(t *testing.T) {
	_, _, err := parseS3Path("s3://")
	assert.EqualError(t, err, "Invalid s3 path s3://")

	_, _, err = parseS3Path("s3://ag-ge")
	assert.EqualError(t, err, "Invalid s3 path s3://ag-ge")
}

func TestFileReader(t *testing.T) {
	// Create a temporary file and write some data to it
	file, err := ioutil.TempFile("/tmp", "pathioFileReaderTest")
	assert.Nil(t, err)
	text := "fileReaderTest"
	ioutil.WriteFile(file.Name(), []byte(text), 0644)

	reader, err := Reader(file.Name())
	assert.Nil(t, err)
	line, _, err := bufio.NewReader(reader).ReadLine()
	assert.Nil(t, err)
	assert.Equal(t, string(line), text)
}

func TestWriteToFilePath(t *testing.T) {
	file, err := ioutil.TempFile("/tmp", "writeToPathTest")
	assert.Nil(t, err)
	defer os.Remove(file.Name())

	assert.Nil(t, Write(file.Name(), []byte("testout")))
	output, err := ioutil.ReadFile(file.Name())
	assert.Nil(t, err)
	assert.Equal(t, "testout", string(output))
}

func TestRegion(t *testing.T) {
	regionObj, err := region("us-west-1")
	assert.Nil(t, err)
	assert.Equal(t, regionObj.EC2Endpoint, "https://ec2.us-west-1.amazonaws.com")

	regionObj, err = region("BadRegion")
	assert.NotNil(t, err)
}
