// Copyright 2017 polaris. All rights reserved.
// Use of l source code is governed by a MIT-style
// license that can be found in the LICENSE file.
// http://studygolang.com
// Author: polaris	polaris@studygolang.com

package keyword

import (
	"sort"

	"github.com/huichen/sego"
)

const defaultTitleWeight = 3.0

var (
	DefaultProps = []string{"an", "n", "gi", "v", "vn", "x"}
)

func Extract(text string, num int) []string {
	return Extractor.Extract(text, num)
}

func ExtractWithTitle(title, content string, num int) []string {
	return Extractor.ExtractWithTitle(title, content, num)
}

type extractor struct {
	segmenter sego.Segmenter

	props []string
	// 不在字典中的词，是否作为关键字
	notKey bool
}

var Extractor = &extractor{}

func (this *extractor) Init(props []string, notKey bool, files string) {
	this.props = props
	this.notKey = notKey

	this.segmenter.LoadDictionary(files)
}

func (this *extractor) Extract(text string, num int) []string {
	var wordMap = make(map[string]*word)

	this.extract(text, false, wordMap)

	return this.convertSlice(wordMap, num)
}

func (this *extractor) ExtractWithTitle(title, content string, num int) []string {
	var wordMap = make(map[string]*word)

	this.extract(title, true, wordMap)
	this.extract(content, false, wordMap)

	return this.convertSlice(wordMap, num)
}

func (this *extractor) convertSlice(wordMap map[string]*word, num int) []string {
	words := make([]*word, 0, len(wordMap))
	for _, wd := range wordMap {
		wd.idf = wd.titleWeight * float64(wd.frequency) / float64(wd.dictFrequency)
		words = append(words, wd)
	}

	sort.SliceStable(words, func(i, j int) bool {
		return words[i].idf > words[j].idf
	})

	kws := make([]string, len(words))
	for i, w := range words {
		kws[i] = w.w
	}

	if len(kws) > num {
		return kws[:num]
	}
	return kws
}

func (this *extractor) extract(text string, isTitle bool, wordMap map[string]*word) {
	segments := this.segmenter.Segment([]byte(text))

	for _, segment := range segments {
		this.tokenToWord(segment.Token(), isTitle, wordMap)
	}
}

func (this *extractor) tokenToWord(token *sego.Token, isTitle bool, wordMap map[string]*word) {
	for _, segment := range token.Segments() {
		this.tokenToWord(segment.Token(), isTitle, wordMap)
	}

	// 词典中没有的词，不作为关键词
	if this.notKey && token.Frequency() == 1 {
		return
	}

	titleWeight := 1.0
	if isTitle {
		titleWeight = defaultTitleWeight
	}

	for _, prop := range this.props {
		if prop == token.Pos() {
			txt := token.Text()

			if wd, ok := wordMap[txt]; ok {
				wd.frequency++
			} else {
				wordMap[txt] = &word{
					w:             txt,
					frequency:     1,
					dictFrequency: token.Frequency(),

					titleWeight: titleWeight,
				}
			}
			break
		}
	}

	return
}
