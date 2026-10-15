Runs the fan of the **Argon ONE V2** case (Raspberry Pi 4) and its power button.

- **Fan:** three steps. From *Fan on at* the fan runs at the first speed, from
  *Faster at* at the second, from *Fastest at* at the third. It slows down
  again 3 °C below a step, so it doesn't switch back and forth.
- **Power button:** press it twice to restart RasBeacon; hold it for 3 seconds
  to shut it down. Holding it for 5 seconds cuts the power at once (the case
  does that by itself).

When the plugin is stopped, the fan keeps its last speed.
