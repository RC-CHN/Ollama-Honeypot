package main

import "time"

var modelList []ListModelResponse = []ListModelResponse{qwq, deepseekR1_14b, deepseekR1_32b, deepseekR1_671b}
var modelNameMap = map[string]struct{}{
	qwq.Name:             {},
	deepseekR1_14b.Name:  {},
	deepseekR1_32b.Name:  {},
	deepseekR1_671b.Name: {},
}
var qwq = ListModelResponse{
	Model:      "qwq:latest",
	Name:       "qwq:latest",
	Size:       int64(19851349390),
	Digest:     "cc1091b0e276012ba4c1662ea103be2c87a1543d2ee435eb5715b37b9b680d27",
	ModifiedAt: time.Date(2025, 3, 7, 9, 59, 48, 0, time.FixedZone("CST", 8*3600)),
	Details: ModelDetails{
		ParentModel:       "",
		Format:            "gguf",
		Family:            "qwen2",
		Families:          []string{"qwen2"},
		ParameterSize:     "32.8B",
		QuantizationLevel: "Q4_K_M",
	},
}

var deepseekR1_32b = ListModelResponse{
	Model:      "deepseek-r1:32b",
	Name:       "deepseek-r1:32b",
	Size:       int64(19851337640),
	Digest:     "38056bbcbb2d068501ecb2d5ea9cea9dd4847465f1ab88c4d4a412a9f7792717",
	ModifiedAt: time.Date(2025, 3, 23, 10, 26, 6, 701448900, time.FixedZone("CST", 8*3600)),
	Details: ModelDetails{
		ParentModel:       "",
		Format:            "gguf",
		Family:            "qwen2",
		Families:          []string{"qwen2"},
		ParameterSize:     "32.8B",
		QuantizationLevel: "Q4_K_M",
	},
}

var deepseekR1_14b = ListModelResponse{
	Model:      "deepseek-r1:14b",
	Name:       "deepseek-r1:14b",
	Size:       int64(8988112040),
	Digest:     "ea35dfe18182f635ee2b214ea30b7520fe1ada68da018f8b395b444b662d4f1a",
	ModifiedAt: time.Date(2025, 3, 20, 16, 50, 20, 450466500, time.FixedZone("CST", 8*3600)),
	Details: ModelDetails{
		ParentModel:       "",
		Format:            "gguf",
		Family:            "qwen2",
		Families:          []string{"qwen2"},
		ParameterSize:     "14.8B",
		QuantizationLevel: "Q4_K_M",
	},
}

var deepseekR1_671b = ListModelResponse{
	Model:      "deepseek-r1:671b",
	Name:       "deepseek-r1:671b",
	Size:       int64(404430188519),
	Digest:     "739e1b229ad7f02d88c5ea4a7d3fda19f7b46170c233024025feeaa6338b9a46",
	ModifiedAt: time.Date(2025, 3, 20, 16, 50, 8, 48004, time.FixedZone("CST", 8*3600)),
	Details: ModelDetails{
		Format:            "gguf",
		Family:            "deepseek2",
		Families:          []string{"deepseek2"},
		ParameterSize:     "671.0B",
		QuantizationLevel: "Q4_K_M",
	},
}
