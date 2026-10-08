package naming

import "strings"

// Ported from swift/lib/Basic/PartsOfSpeech.def and StringExtras.cpp,
// commit f79ab15f11ffea6cad859da22b17da6cd9252167. Modified for Go lookup.
// Copyright (c) 2014 - 2017 Apple Inc. and the Swift project authors.
// Apache-2.0 WITH Swift-exception; see THIRD_PARTY_NOTICES.md.
var prepositions = wordSet(
	"above after along alongside as at before below by following for from given in including inside " +
		"into matching of on passing preceding since to until using via when with within",
)
var verbs = wordSet(
	"abbreviate accept activate add adjust admire admit advise afford agree alert allow alter amuse " +
		"analyse analyze animate announce annoy answer apologise appear append applaud apply apportion " +
		"appreciate approve argue arrange arrest arrive ask assign attach attack attempt attend attract " +
		"avoid awake back bake balance ban bang bare bat bathe battle be beat become beg begin behave " +
		"belong bend bet bid bite bleach bless blind blink blot blow blush boast boil bolt bomb book bore " +
		"borrow bounce bow box brake branch break breathe bring broadcast bruise brush bubble build bump " +
		"burn bury buy buzz calculate call camp cancel capture care carry carve cast catch cause center " +
		"challenge change charge chase cheat check cheer chew choke choose chop claim clap clean clear " +
		"click close coach coil collect collapse colour comb come command commit communicate compare " +
		"compete complain complete concentrate concern confess confuse connect consider consist contain " +
		"contains continue convert copy correct cough cost count cover crack crash crawl cross crush cry " +
		"cure curl curve customize cut cycle dam damage dance dare decay deceive decide decode decorate " +
		"defer define delay delete delight deliver depend describe deselect desert deserve destroy detach " +
		"detect develop dig dim disagree disappear disapprove disarm discover dislike dismiss display " +
		"divide do double doubt drag drain draw dream dress drink drip drive drop drown drum dry " +
		"duplicate dust earn eat echo edit educate embarrass employ empty enable encode encourage end " +
		"enjoy enter entertain enumerate enqueue escape examine excite excuse execute exercise exist " +
		"expand expect explain explode export extend face fade fail fancy fasten fax fear feel fence " +
		"fetch fight fill film find finish fire fit fix flap flash flatten flip float flood flow flower " +
		"fly focus fold follow fool force forget forgive form found freeze frighten fry gain gather gaze " +
		"generate get give glow glue go grab grate grease greet grin grip groan grow guarantee guard " +
		"guess guide hammer hand handle hang happen harass harm hate haunt head heal heap hear heat help " +
		"hide highlight hit hold hook hop hope hover hug hum hunt hurry hurt identify ignore imagine " +
		"import impress improve include increase influence inform inject injure insert instruct intend " +
		"interest interfere interrupt intersect intersects introduce invent invite irritate itch jail jam " +
		"jog join joke judge juggle jump keep kick kill kiss kneel knit knock knot know label land last " +
		"laugh launch lay lead learn leave lend let level license lick lie lighten like listen live load " +
		"localize lock long look lose love maintain make man manage march mark marry match mate matter " +
		"mean measure meddle meet melt memorise mend merge mess milk mine miss minus mix moan moor mourn " +
		"move muddle mug multiply murder nail nest nod normalize note notice notify number obey observe " +
		"obtain occur offend offer open order overflow owe own pack paddle paint park part pass paste pat " +
		"pause pay peck pedal peel peep perform permit phone pick pinch pine place plan plant play please " +
		"plug poke polish pop possess post pour practice practise pray preach precede prefer preload " +
		"prepare prepend present preserve press pretend prevent prick print produce program promise " +
		"protect provide pull pump punch puncture punish push put question queue race radiate rain raise " +
		"reach read realise receive recognise record reduce reflect refuse register regret reign reject " +
		"rejoice relax release rely remain remember remind remove repair repeat replace reply report " +
		"request require resize rescue resolve retain retire return reverse review rhyme ride ring rinse " +
		"rise risk rob rock roll rot rub ruin rule run rush sack sail satisfy save saw say scale scare " +
		"scatter scold scorch scrape scratch scream screw scribble scroll scrub seal search see select " +
		"sell send separate serve settle shade share shave shelter shiver shock shop show shrug shut sigh " +
		"sign signal sin sing sip sit ski skip slap sleep slip slow smash smell smile smoke snatch sneeze " +
		"sniff snore snow soak soothe sound spare spark sparkle speak spell spend spill spoil spot spray " +
		"sprout squash squeak squeal squeeze stain stamp stand standardise standardize stare start stay " +
		"steer step stir stitch stop store strap strengthen stretch strip stroke stuff subtract succeed " +
		"suck suffer suggest suit supply support suppose suppress surprise surround suspect suspend swim " +
		"switch take talk tame tap taste teach tear tease telephone tell tempt terrify test thank thaw " +
		"think throw tick tickle tie time tip tire toggle touch tour tow trace trade train translate " +
		"transform transport trap travel traverse treat tremble trick trip trot trouble truncate trust " +
		"try tug tumble turn twist understand undress unfasten union unite unload unlock unpack untidy up " +
		"update use validate vanish visit wail wait wake walk wander want warm warn wash waste watch " +
		"water wave wear weigh welcome whine whip whirl whisper whistle win wink wipe wish wobble wonder " +
		"work worry wrap wreck wrestle wriggle write yawn yell zip zoom",
)

func wordSet(s string) map[string]bool {
	result := map[string]bool{}
	for _, w := range strings.Fields(s) {
		result[w] = true
	}
	return result
}
func speech(s string) string {
	s = strings.ToLower(s)
	if prepositions[s] {
		return "preposition"
	}
	if verbs[s] {
		return "verb"
	}
	if strings.HasSuffix(s, "ing") && len(s) > 4 {
		root := strings.TrimSuffix(s, "ing")
		if speech(root) == "verb" || (!strings.HasSuffix(root, "e") && speech(root+"e") == "verb") || (len(root) > 1 && root[len(root)-1] == root[len(root)-2] && speech(root[:len(root)-1]) == "verb") {
			return "gerund"
		}
	}
	for _, prefix := range []string{"auto", "re", "de"} {
		if len(s) > len(prefix) && strings.HasPrefix(s, prefix) && speech(s[len(prefix):]) == "verb" {
			return "verb"
		}
	}
	return ""
}
