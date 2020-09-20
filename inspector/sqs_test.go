package inspector

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestExtractRegionFromURL(t *testing.T) {
	var region *string
	var err error

	// Valid case
	region, err = extractRegionFromURL("https://sqs.ap-northeast-1.amazonaws.com/123456789xxx/attribute-queue")
	assert.NoError(t, err)
	assert.Equal(t, "ap-northeast-1", *region)

	// Invalid cases
	region, err = extractRegionFromURL("https://sns.ap-northeast-1.amazonaws.com/123456789xxx/attribute-queue")
	assert.Error(t, err)
	assert.Nil(t, region)

	region, err = extractRegionFromURL("https://sqs.ap-northeast-1.xxx.amazonaws.com/123456789xxx/attribute-queue")
	assert.Error(t, err)
	assert.Nil(t, region)
}
