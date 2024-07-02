// From what i understand, the inbuilt json.Marshal and json.Unmarshal functions
// are slow, so we use a third party library called json-iterator, which is
// an optimized version of the inbuilt version. As we're doing a lot of
// unmarshalling for the api responses in the poolers, using a faster
// version of it might make stuff more efficient.
//
// https://github.com/json-iterator/go
package encoder

import jsoniter "github.com/json-iterator/go"

var JSON = jsoniter.ConfigCompatibleWithStandardLibrary
