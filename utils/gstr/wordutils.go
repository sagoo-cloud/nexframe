package gstr

import (
	"sort"
	"strings"
	"sync"
	"time"
	"unicode/utf8"
)

// WordRankResult 表示单词频率分析的结果
type WordRankResult struct {
	Text  string  `json:"text"`
	Count int     `json:"count"`
	Rank  float32 `json:"rank"`
}

// WordRankResults 是 WordRankResult 的切片，实现了 sort.Interface
type WordRankResults []WordRankResult

func (s WordRankResults) Len() int           { return len(s) }
func (s WordRankResults) Less(i, j int) bool { return s[i].Count > s[j].Count }
func (s WordRankResults) Swap(i, j int)      { s[i], s[j] = s[j], s[i] }

// cacheItem 表示缓存中的一个项目
type cacheItem struct {
	value      WordRankResults
	expiration time.Time
}

var (
	// 使用 sync.Pool 来重用 map 对象
	wordCountPool = sync.Pool{
		New: func() interface{} {
			return make(map[string]int)
		},
	}

	// 使用 sync.Map 来缓存频繁调用的结果
	wordRankCache sync.Map

	// 缓存过期时间
	cacheDuration = 5 * time.Minute
)

// WordRank 对输入的单词切片进行词频分析
// 参数：
//   - arrWords: 单词切片
//
// 返回值：按词频排序的 WordRankResults
func WordRank(arrWords []string) WordRankResults {
	// 生成缓存键
	cacheKey := strings.Join(arrWords, ",")

	// 检查缓存中是否存在结果
	if cachedResult, found := wordRankCache.Load(cacheKey); found {
		item := cachedResult.(cacheItem)
		if time.Now().Before(item.expiration) {
			return item.value
		}
		// 如果缓存已过期，删除它
		wordRankCache.Delete(cacheKey)
	}

	// 从对象池获取 map
	wordCount := wordCountPool.Get().(map[string]int)
	defer func() {
		// 清空 map 并放回对象池
		for k := range wordCount {
			delete(wordCount, k)
		}
		wordCountPool.Put(wordCount)
	}()

	totalWords := 0

	for _, word := range arrWords {
		word = strings.TrimSpace(word)
		if word != "" && utf8.RuneCountInString(word) > 1 {
			wordCount[word]++
			totalWords++
		}
	}

	result := make(WordRankResults, 0, len(wordCount))
	for word, count := range wordCount {
		result = append(result, WordRankResult{
			Text:  word,
			Count: count,
			Rank:  float32(count) / float32(totalWords),
		})
	}

	sort.Sort(result)

	// 将结果存入缓存
	wordRankCache.Store(cacheKey, cacheItem{
		value:      result,
		expiration: time.Now().Add(cacheDuration),
	})

	return result
}

// ClearWordRankCache 清除 WordRank 函数的缓存
func ClearWordRankCache() {
	wordRankCache.Range(func(key, value interface{}) bool {
		wordRankCache.Delete(key)
		return true
	})
}

// SetCacheDuration 设置缓存的过期时间
func SetCacheDuration(duration time.Duration) {
	cacheDuration = duration
}
