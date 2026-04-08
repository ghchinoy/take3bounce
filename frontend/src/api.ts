/**
 * Copyright 2026 Google LLC
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

export interface VoiceActor {
  shortName: string;
  baseVoice: string;
  stylePrompt: string;
}

export interface Variation {
  take: string;
  persona: string;
  subtext: string;
  technicalEnergy: string;
  text: string;
  audio?: string;
  mimeType?: string;
}

/**
 * generateVariations triggers the backend to orchestrate three distinct
 * audio takes (Safe, Pushed, Wildcard) based on the provided text and persona.
 */
export async function generateVariations(text: string, voiceActor: VoiceActor): Promise<Variation[]> {
  const response = await fetch('/api/variations', {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify({ text, voiceActor })
  });

  if (!response.ok) {
    const errText = await response.text();
    throw new Error(`Variations failed: ${errText}`);
  }

  return response.json();
}

/**
 * retryAudio triggers the backend to regenerate a single specific take.
 * This is used for bypassing PROHIBITED_CONTENT blocks or manually regenerating
 * audio after a user edits the tags in the UI.
 */
export async function retryAudio(variation: Variation, voiceActor: VoiceActor): Promise<Variation> {
  const response = await fetch('/api/retry-audio', {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify({ variation, voiceActor })
  });

  if (!response.ok) {
    const errText = await response.text();
    throw new Error(`Retry failed: ${errText}`);
  }

  return response.json();
}
