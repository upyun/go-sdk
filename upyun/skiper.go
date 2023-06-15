package upyun

// 创建一个map，用于加速获取已经上传完成的ID
func makeSkipPartIds(records []*DisorderPart) map[int64]int64 {
	result := make(map[int64]int64)
	if records == nil {
		return result
	}

	for _, v := range records {
		result[v.ID] = v.Size
	}
	return result
}

// 分片跳过器，用于跳过已经上传的分片
type skiper struct {
	nextPartId int64
	parts      map[int64]int64
}

func createSkiper(nextPartId int64, disorderParts []*DisorderPart) *skiper {
	return &skiper{
		nextPartId: nextPartId,
		parts:      makeSkipPartIds(disorderParts),
	}
}

// 找到最早缺失的 partId
func (p *skiper) FirstMissPartId() int64 {
	if p.nextPartId == 0 && len(p.parts) == 0 {
		return 0
	}

	if len(p.parts) > 0 {
		var i int64
		for {
			if _, ok := p.parts[i]; !ok {
				return i
			}
			i++
		}
	}
	return p.nextPartId
}

// 判断 partId 是否已经被上传 需要被跳过
func (p *skiper) IsSkip(partId int64) bool {
	if p.nextPartId > partId {
		return true
	}
	if _, ok := p.parts[partId]; ok {
		return true
	}
	return false
}
