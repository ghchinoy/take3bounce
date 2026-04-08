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
