+++
author = "Antti Viitala"
title = "Anti-aliasing for Garmin watch faces"
date = "2023-01-19"
description = ""
tags = [
    "development"
]
thumbnail = "/content/antialias-comparison.png" ## UPDATE?
images = ["/content/antialias-comparison.png"] ## UPDATE?
series = []
+++

## What is anti-aliasing?

Anti-aliasing is a technique used to smoothen the edges of a rendered shape on a computer screen. Since screens are just large grids of pixels, rendering rounded, smooth shapes can look "jagged" when the contrast between the background (e.g. white) and foreground colors (e.g. black) is large. Anti-aliasing helps counter this by adding gray pixels in the background-foreground boundary, which makes the output look more attractive to the human eye.

## Enabling anti-aliasing for your watch face

When I first started working with watch faces, I actually looked for an option to do something like this but couldn't find a solution at the time. However, since then it has become [very simple to do](https://developer.garmin.com/connect-iq/core-topics/graphics/):

```c
// Within the onUpdate() function
...
    // Enable anti-aliasing, if available
    if(dc has :setAntiAlias) {
        dc.setAntiAlias(true);
    }
    // Rest of your watch face code
...
```

Here's a side-by-side comparison, you can probably tell which one is before/after!

![comparison](/content/antialias-comparison.png)
