// Copyright 2026 Google LLC
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package main

// FullTagsList defines the uncurated "Full Firehose" audio tag vocabulary injected
// into the director prompt during A/B prompt strategy selection.
const FullTagsList = `You may use any of the following tags to shape the performance:

[acceptance], [accomplishment], [achievement], [action], [admiration], [admonition], [adoration], [aesthetic appreciation], [affection], [aggression], [agitation], [alarm], [amazed], [ambivalence], [amused], [amusement], [analysis], [anger], [animation], [annoyance], [anticipation], [anxiety], [apology], [appreciation], [apprehension], [approval], [arrogance], [assertion], [assertive], [assertiveness], [assurance], [asterisk], [astonishment], [aversion], [awareness], [awe], [awkwardness], [bargaining], [boredom], [caring], [caution], [cautious], [certainty], [challenging], [comfort], [compassion], [concentration], [concern], [confidence], [confident], [confident/neutral], [confusion], [contemplative], [contempt], [contentment], [conviction], [courage], [craving], [critical], [criticism], [curiosity], [decision], [defiance], [demonstration], [description], [descriptive], [desire], [despair], [desperation], [despondency], [determination], [determined], [devotion], [directness], [disagreement], [disappointment], [disapproval], [disbelief], [discernment], [discomfort], [disdain], [disgust], [disillusionment], [dislike], [dismissal], [dismissive], [dismissiveness], [distress], [doubt], [dread], [eagerness], [effervescence], [embarrassment], [embitterment], [embracement], [empathy], [emphasis], [enchantment], [encouragement], [encouraging], [energetic], [energy], [enjoyment], [enthusiasm], [enthusiastic], [excited], [exhaustion], [explaining], [explanation], [explanatory], [fascination], [fear], [focus], [fondness], [friendliness], [friendly], [frustration], [gratification], [gratitude], [grief], [guilt], [happiness], [happy], [high energy/active], [hope], [horror], [humor], [hurt], [incredulity], [indifference], [indignation], [information], [information disclosure], [informational], [informative], [informativeness], [instruction], [interest], [intrigue], [invitation], [joy], [logical reasoning], [love], [low energy/passive], [melancholy], [mixed], [musement], [negative], [negative surprise], [nervousness], [neutral], [none], [nostalgia], [observation], [offense], [optimism], [pain], [panic], [passion], [pensive], [pessimism], [pity], [planning], [playful], [pleading], [pleased], [positive], [positive surprise], [praise], [pride], [realization], [recognition], [reflection], [regret], [relaxation], [relief], [reminiscence], [resignation], [sadness], [sarcastic], [satisfaction], [self-deprecation], [self-satisfaction], [sentimentality], [serenity], [seriousness], [shame], [shock], [skepticism], [slight relief], [smitten], [solemnity], [speculation], [spk1], [strategizing], [stress], [struggle], [success], [suffering], [suggestion], [summary], [surprise], [suspicion], [sympathy], [tension], [terror], [thanks], [thinking], [thoughtful], [thoughtfulness], [thrill], [tiredness], [triumph], [uncertainty], [unclear], [understanding], [unease], [urgency], [victory], [warning], [weariness], [wisdom], [wistful], [worry], [yearning], [sighs], [laughs], [uhm], [short pause], [medium pause], [long pause], [robotic], [shouting], [whispers], [extremely fast], [fast], [slow]`

// EnhancedTagsList defines the categorized audio tag vocabulary injected into
// the director prompt during A/B prompt strategy selection.
const EnhancedTagsList = `You may use tags from the following curated categories to shape the performance:

1. Non-Speech Sounds: [sighs], [laughs], [uhm], [gasp], [crying], [giggles], [snorts]
2. Style Modifiers: [sarcastic], [robotic], [shouting], [whispers], [extremely fast], [trembling]
3. Pacing and Pauses: [short pause], [medium pause], [long pause]
4. Core Emotional Tones: [joy], [grief], [anger], [fear], [surprise], [disgust], [anticipation], [trust]
5. Nuanced Deliveries: [scared], [curious], [bored], [mischievously], [panicked], [serious], [tired], [amusement], [confusion], [despair], [excited], [exhaustion], [frustration], [hope], [love], [nostalgia], [pride], [relief], [serenity], [shame], [sympathy], [terror], [wisdom], [wistful], [confident], [informative], [playful], [neutral]`
