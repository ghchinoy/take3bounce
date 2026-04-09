export interface Preset {
  label: string;
  texts: string[];
}

export const PRESETS: Preset[] = [
  {
    label: "Kittens",
    texts: [
      "Our kittens are raised in a cage-free environment with 24/7 medical supervision.",
      "Every kitten goes home with a full starter kit, including their favorite toys and premium food.",
      "We believe that a happy kitten makes a happy home, which is why we focus on early socialization."
    ]
  },
  {
    label: "Hamlet",
    texts: [
      "To be, or not to be, that is the question: Whether 'tis nobler in the mind to suffer the slings and arrows of outrageous fortune, or to take arms against a sea of troubles and by opposing end them.",
      "There is nothing either good or bad, but thinking makes it so.",
      "This above all: to thine own self be true, And it must follow, as the night the day, Thou canst not then be false to any man."
    ]
  },
  {
    label: "Wuthering Heights",
    texts: [
      "I lingered round them, under that benign sky; watched the moths fluttering among the heath and harebells, listened to the soft wind breathing through the grass, and wondered how any one could ever imagine unquiet slumbers for the sleepers in that quiet earth.",
      "He's more myself than I am. Whatever our souls are made of, his and mine are the same.",
      "If all else perished, and he remained, I should still continue to be; and if all else remained, and he were annihilated, the universe would turn to a mighty stranger."
    ]
  }
];
