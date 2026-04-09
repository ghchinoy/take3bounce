export interface Quote {
  text: string;
  attribution?: string;
}

export interface Preset {
  label: string;
  texts: Quote[];
}

export const PRESETS: Preset[] = [
  {
    label: "Kittens",
    texts: [
      { text: "Our kittens are raised in a cage-free environment with 24/7 medical supervision." },
      { text: "Every kitten goes home with a full starter kit, including their favorite toys and premium food." },
      { text: "We believe that a happy kitten makes a happy home, which is why we focus on early socialization." }
    ]
  },
  {
    label: "Hamlet",
    texts: [
      { text: "To be, or not to be, that is the question: Whether 'tis nobler in the mind to suffer the slings and arrows of outrageous fortune, or to take arms against a sea of troubles and by opposing end them.", attribution: "William Shakespeare" },
      { text: "There is nothing either good or bad, but thinking makes it so.", attribution: "William Shakespeare" },
      { text: "This above all: to thine own self be true, And it must follow, as the night the day, Thou canst not then be false to any man.", attribution: "William Shakespeare" }
    ]
  },
  {
    label: "Wuthering Heights",
    texts: [
      { text: "I lingered round them, under that benign sky; watched the moths fluttering among the heath and harebells, listened to the soft wind breathing through the grass, and wondered how any one could ever imagine unquiet slumbers for the sleepers in that quiet earth.", attribution: "Emily Brontë" },
      { text: "He's more myself than I am. Whatever our souls are made of, his and mine are the same.", attribution: "Emily Brontë" },
      { text: "If all else perished, and he remained, I should still continue to be; and if all else remained, and he were annihilated, the universe would turn to a mighty stranger.", attribution: "Emily Brontë" },
      { text: "My love for Linton is like the foliage in the woods: time will change it, I’m well aware, as winter changes the trees. My love for Heathcliff resembles the eternal rocks beneath: a source of little visible delight, but necessary. Nelly, I am Heathcliff! He’s always, always in my mind: not as a pleasure, any more than I am always a pleasure to myself, but as my own being.", attribution: "Emily Brontë" }
    ]
  },
  {
    label: "Inspiring",
    texts: [
      { text: "We choose to go to the moon in this decade and do the other things, not because they are easy, but because they are hard. Because that goal will serve to organize and measure the best of our energies and skills.", attribution: "John F. Kennedy" },
      { text: "Earth is the cradle of humanity, but one cannot live in a cradle forever. The pursuit of light and space will lead us to penetrate the bounds of the atmosphere, timidly at first, but in the end to conquer the whole of solar space.", attribution: "Konstantin Tsiolkovsky" },
      { text: "We shall go to the moon, we shall go to the planets, we shall travel to the stars just as today we go from Liverpool to New York. The oceans of space will be crossed like the seas of the moon!", attribution: "Jules Verne, From the Earth to the Moon (1865)" },
      { text: "Houston, Tranquility Base here. The Eagle has landed. It’s one small step for a man, one giant leap for mankind. Looking back at the Earth, it is a magnificent desolation.", attribution: "Neil Armstrong & Buzz Aldrin (NASA Historic Archive)" },
      { text: "Marvel not, my comrade, if I appear talking to you on super-terrestrial and aerial topics. The long and the short of the matter is that I am running over the order of a journey I have lately made to the moon.", attribution: "Lucian of Samosata, 2nd Century AD (A True Story)" },
      { text: "We don't build for the quiet days. We build for the moments when the pressure spikes, the lights flicker, and the room goes completely silent. Because playing it safe never changed the world. We are the architects of 'what's next.' And we're just getting started." }
    ]
  }
];