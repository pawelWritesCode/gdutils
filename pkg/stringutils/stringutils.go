// Package stringutils holds utility methods for working with strings.
package stringutils

import (
	"math/rand"
	"time"
)

const (
	// CharsetUnicode represents set of Unicode characters (contain multi byte runes).
	CharsetUnicode = CharsetEmoji + CharsetSex + CharsetKeyboard + CharsetCurrencies + CharsetGreek +
		CharsetMathematicsSymbols + CharsetMathematicsFonts + CharsetInvertedLetters + CharsetBraille +
		CharsetIPA + CharsetFullWidthCharacters + CharsetUnits + CharsetLatin + CharsetCyrillic + CharsetChinese +
		CharsetJapanese + CharsetKorean + CharsetArabic + CharsetEthiopian + CharsetDevanagari + CharsetBengali +
		CharsetTamil + CharsetTibetan + CharsetHieroglyphs + CharsetCuneiform + CharsetLinearB + CharsetPhoenician +
		CharsetRunes + CharsetDesered + CharsetShavian

	// CharsetASCII represents set containing only ASCII characters.
	CharsetASCII = " !#$%&'()*+,-.0123456789:;=?@ABCDEFGHIJKLMNOPQRSTUVWXYZ^_`abcdefghijklmnopqrstuvwxyz|~"

	// CharsetPolish represents set of only polish letters.
	CharsetPolish = "ĄąĆćĘęŁłŃńÓóŚśŹźŻżabcdefghijklmnoprstuwvxyzABCDEFGHIJKLMNOPRSTUWVXYZ"

	// CharsetEnglish represents set of only english letters.
	CharsetEnglish = "abcdefghijklmnoprstuwvxyzABCDEFGHIJKLMNOPRSTUWVXYZ"

	// CharsetRussian represents set of only russian letters.
	CharsetRussian = "АаБбВвГгДдЕеЁёЖжЗзИиЙйКкЛлМмНнОоПпРрСсТтУуФфХхЦцЧчШшЩщЪъЫыЬьЭэЮюЯя"

	// CharsetMathematicsSymbols represents set of mathematical symbols.
	CharsetMathematicsSymbols = "∛ℤℝ⋃⋂⋡⪺≠⊐∵∧⦝⊕∯∑∏⊶⫝̸⦕⟪⦅」⟅⦋"

	// CharsetMathematicsFonts represents set of different letters and numbers in different mathematical fonts.
	CharsetMathematicsFonts = "𝔄𝔞𝕭𝖇𝔻𝕕𝟛𝐅𝐟𝐿𝑙𝑴𝒎𝒩𝓃𝓞𝓸𝖶𝗐𝗫𝘅𝟭𝘠𝘺𝚉𝚣"

	// CharsetEmoji represents set of emojis
	CharsetEmoji = "🤡🤖🧟🏋🥇☟💄🐲🌓🌪🇵🇱💵"

	// CharsetCurrencies represents set of financial currencies.
	CharsetCurrencies = "₿£$¢₰₱₣"

	// CharsetSex represents set of symbols related with genders.
	CharsetSex = "⚥⚩⚮⚭"

	// CharsetKeyboard represents set of keyboard symbols.
	CharsetKeyboard = "⌘⎈✄↵✧✲"

	// CharsetGreek represents set of greek characters.
	CharsetGreek = "αβδεθλμπφψΩ"

	// CharsetInvertedLetters represents set of inverted letters.
	CharsetInvertedLetters = "ɐqɔԀÒᴚS"

	// CharsetIPA International Phonetic Alphabet (IPA) is a phonetic (pronunciation) notation designed for all human languages.
	CharsetIPA = "āäēĕæɛiːɝo͞o"

	// CharsetFullWidthCharacters A full-width character means the character has the same width as a Chinese character,
	// regardless of font choice.
	CharsetFullWidthCharacters = "０ＣＤ￦￤ｒｐａｒｔｉｃｕｌａｒ"

	// CharsetUnits represents set of unit symbols
	CharsetUnits = "㎛㎠㎢㎖㎲㎏㎆㎑㎷㏀㎃㎨㎮㎪㎉㏑㍱㏖"

	// CharsetBraille is used by blind people
	CharsetBraille = "⠞⠽⠏⠑⠓⠑⠗⠑"

	// CharsetLatin Latin Alphabet (aka Roman alphabet) is used to write most Western languages, including:
	// English, German, French, Spanish, Italian.
	CharsetLatin = "ᴁèóƃᶐɷȷᴂɒᴝᴥ"

	// CharsetCyrillic The Cyrillic script is a writing system used for various alphabets used in Eastern Europe,
	// the Caucasus, Central Asia, and North Asia.
	CharsetCyrillic = CharsetRussian + "ԆҤ҂ԔԙӜԨԬӦӴѠѼ"

	// CharsetChinese There are many thousands of Chinese characters. A college educated Chinese person typically know
	// 5 thousand character. 3 thousand is minimum to read newspapers.
	CharsetChinese = "的一是在不了有和人这中大为上个国相见欢·林花谢了春红"

	// CharsetJapanese represents set of Japanese characters.
	CharsetJapanese = "あいうえアイウエオ・ヽヾヿｱｲｳｴｵｶｷｸ"

	// CharsetKorean represents set of Korean characters.
	CharsetKorean = "ᅒᅓᅔᅕᅖᅗᅘᆪᆫᆬᆭᆮᆯᆰ"

	// CharsetArabic represents set of Arabic characters.
	CharsetArabic = "ﷺ﷽ﵴﴙﲀ۝ﲂ۞؊٩١۲ݶمڮجݗݨݳھڝ"

	// CharsetEthiopian represents set of Ethiopian characters. It is third in Africa, followed by West African Cats and East Africa.
	CharsetEthiopian = "ቷቸቹጟጠኸኹⶹꬨꬩꬖꬠ᎐᎑᎒፣፤፥፪፫፬የኢትዮጵያ፡መ"

	// CharsetDevanagari represents set of Indian characters. The most popular language of the Indo-Aryan family is Hindi,
	// and Hindi uses Devanagari as its writing system
	CharsetDevanagari = "ऒओॺॻ॥॰षस१२३ळऴ"

	// CharsetBengali represents set of Bengali characters. Bengali script is mainly used by languages Bengali and Assamese, in India.
	CharsetBengali = "ঊঋজঝঞলশ৪৫৵৶ঁংৗৢ"

	// CharsetTamil represents set of Tamil characters. It is an abugida script that is used by Tamils and Tamil speakers
	// in India, Sri Lanka, Malaysia, Singapore, Indonesia.
	CharsetTamil = "ஃஅஆஇணதந௫௬௭ௐ௳௴ோ ௌ"

	// CharsetTibetan represents set of Tibetan characters.
	CharsetTibetan = "གྷངཅ༳༪༫༐༑༒࿄࿅࿇࿈༚༛༜࿐࿑࿒ ཱ ི ཱི༻༼ ༽ྨ ྩ"

	// CharsetHieroglyphs represents set of ancient egyptian hieroglyphs.
	CharsetHieroglyphs = "𓀆𓀐𓀛𓀬𓀮𓁂𓁈𓐒𓐗𓐬𓁔𓁪𓂐𓂨𓂫𓂺𓃏𓃗𓃾𓄅𓄚𓄥𓅄𓆥𓆫𓆷𓇈𓇙𓈣𓉓𓉶𓊚𓌁𓌬𓍭𓎒𓏈𓏱"

	// CharsetCuneiform represents set of Cuneiform characters. t was originally used for the Sumerian language,
	// later also used for Semitic Akkadian (Assyrian/Babylonian), Eblaite, Amorite, Elamite, Hattic, Hurrian, Urartian, Hittite, Luwian.
	CharsetCuneiform = "𒀂𒀃𒀄𒁏𒂓𒃶𒄬"

	// CharsetLinearB represents set of LinearB characters. Linear B was used for writing Mycenaean Greek about 1450 BC.
	// It is descended from the older Linear A, an undeciphered earlier script.
	CharsetLinearB = "𐀀𐀁𐀂𐁂𐁃𐁄𐁐𐁑𐁒𐂂𐂃𐂄𐃨𐃩𐃪𐘃𐘄𐘅𐙁𐙂𐙃"

	// CharsetPhoenician represents set of Phoenican characters. Phoenician is the oldest alphabet. It began around 1200 BC.
	CharsetPhoenician = "𐤟𐤛𐤗𐤘𐤒𐤓𐤔𐤕"

	// CharsetRunes represents set of old runes. Rune alphabets was used to write various
	// Germanic languages before the adoption of the Latin alphabet.
	CharsetRunes = "ᚠᚡᚢᛋᛌᛍᛴᛵᛱᛲ"

	// CharsetDesered represents set of Desered characters. Deseret alphabet is designed to be a simple phonetic alphabet for English.
	CharsetDesered = "𐐀𐐁𐐂𐐆𐐇𐐈𐐖𐐗𐐘𐐸𐐹"

	// CharsetShavian represents set of Shavian characters. Shavian alphabet is designed to be a simple phonetic alphabet for English.
	CharsetShavian = "𐑐𐑑𐑒𐑟𐑠𐑡𐑥𐑦𐑧𐑨"
)

var seededRand *rand.Rand = rand.New(
	rand.NewSource(time.Now().UnixNano()))

// RunesFromCharset returns random slice of runes of given length.
// Argument length indices length of output slice.
// Argument charset indices input charset from which output slice will be composed.
func RunesFromCharset(length int, charset []rune) []rune {
	output := make([]rune, 0, length)
	charsetR := charset

	for i := 0; i < length; i++ {
		output = append(output, charsetR[seededRand.Intn(len(charsetR))])
	}

	return output
}
