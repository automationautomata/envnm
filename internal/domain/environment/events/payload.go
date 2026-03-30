package events

import "encoding/json"

type payloadMap map[string]any

func (p payloadMap) MarshalJSON() ([]byte, error) { return json.Marshal(map[string]any(p)) }

type payloadArray []any

func (p payloadArray) MarshalJSON() ([]byte, error) { return json.Marshal([]any(p)) }
