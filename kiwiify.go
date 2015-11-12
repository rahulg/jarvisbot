package jarvisbot

import (
	"regexp"
)

var re_word *regexp.Regexp
var re_no_mangle *regexp.Regexp
var re_vowels *regexp.Regexp
var re_vowels_inner *regexp.Regexp
var re_UU *regexp.Regexp
var re_uu *regexp.Regexp
var re_non_word *regexp.Regexp

func init() {
	re_word = regexp.MustCompile(`(?m)(?:\A|^|\s)(\S+)`)
	re_no_mangle = regexp.MustCompile(`(\w+://.*|\@\w+)`)
	re_vowels_inner = regexp.MustCompile(`(\S)?([AEIOUaeiou])[AEIOUaeiou]?(\S)?`)
	re_vowels = regexp.MustCompile(`(\S?(?:[AEIOUaeiou][AEIOUaeiou]?)\S?)`)
	re_UU = regexp.MustCompile(`(U[Uu])`)
	re_uu = regexp.MustCompile(`(u[Uu])`)
	re_non_word = regexp.MustCompile(`\W`)
}

func kiwiify_is_boundary(s string) bool {
	return s == "" || re_non_word.MatchString(s)
}

func kiwiify_vowel(vowel_triplet string) string {

	tokens := re_vowels_inner.FindAllStringSubmatch(vowel_triplet, -1)
	if len(tokens) == 0 {
		return vowel_triplet
	}

	pre := tokens[0][1]
	vowel := tokens[0][2]
	post := tokens[0][3]

	if pre == "" && post == "" {
		return vowel_triplet
	}

	new_vowel := ""

	switch vowel {
	case "I":
		if kiwiify_is_boundary(pre) {
			return vowel_triplet
		}
		new_vowel = "U"
	case "i":
		if kiwiify_is_boundary(pre) {
			return vowel_triplet
		}
		new_vowel = "u"
	case "E":
		if kiwiify_is_boundary(post) {
			return vowel_triplet
		}
		new_vowel = "I"
	case "e":
		if kiwiify_is_boundary(post) {
			return vowel_triplet
		}
		new_vowel = "i"
	case "A":
		new_vowel = "E"
	case "a":
		new_vowel = "e"
	case "O":
		new_vowel = "U"
	case "o":
		new_vowel = "u"
	default:
		return vowel_triplet
	}

	return pre + new_vowel + post

}

func kiwiify_word(word string) string {
	if re_no_mangle.MatchString(word) {
		return word
	} else {
		word = re_vowels.ReplaceAllStringFunc(word, kiwiify_vowel)
		word = re_UU.ReplaceAllLiteralString(word, "U")
		word = re_uu.ReplaceAllLiteralString(word, "u")
		return word
	}
}

func Kiwiify(msg string) string {
	return re_word.ReplaceAllStringFunc(msg, kiwiify_word)
}
