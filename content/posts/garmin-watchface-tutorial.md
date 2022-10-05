+++
author = "Antti Viitala"
title = "Making a custom Garmin watch face"
date = "2022-10-04"
description = "Use Visual Studio Code and the Garmin SDK to build a custom digital watch face and release it to the Garmin ConnectIQ store."
tags = [
    "development"
]
+++

![design](/content/header-small-1.png)

<!-- 
1. Picture of my shitty drawing
2. Picture in Garmin simulator
3. Picture in Garmin store
4. Picture of Fenix 6 running the watch face

Links:
- Repository link
- Connect IQ link to watch face
-->

## High-level view of the key steps

1. (Assumed as pre-requisite): Install [Visual Studio Code](https://code.visualstudio.com/)
1. (Assumed as pre-requisite): Have a Garmin account
1. [Installing the Garmin Connect SDK](#environment-and-sdk-setup-for-development)
1. [Installing the MonkeyC extension (Garmin SDK extension for project support, development and code analysis)](#environment-and-sdk-setup-for-development)
1. [Set up a new MonkeyC Watch face project](#starting-a-project)
1. [Create the digital time display](#rendering-time-in-digital-format)
1. [Create a function to draw 'gauges'](#creating-data-gauges)
1. [Configure data gauges](#creating-data-gauges)
1. [Compile the project](#compiling-the-project)
1. [Upload to Garmin Connect IQ store](#uploading-to-garmin-connect-iq-store)
1. [Download the watch face to your watch](#downloading-the-watch-face-to-your-watch)

## Design

The design for this watch face is very simple and prioritizes legibility of a small number of key metrics that I care about:

* **Current time**: Digital 24-hour format, hours and minutes only, without separator for hours/minutes - e.g. 1527 for 3:27pm.
* **Current date**: Without year, in dd-mm format - e.g. 31-12 for 31st of December
* **Battery**:
  * Overall battery %
  * Estimated GPS battery life for a GPS-enabled activity
* **Sunrise/sunset**: Show 'next' value for both in hhmm format - e.g. 0612 for 6:12am
* **Steps**: Show today's steps versus a pre-defined (here hardcoded) target - e.g. 10k

Based on the above, I created a rough sketch of what to develop, shown below. I added the 'hour' indices to help with placement of the gauges, they will not be render as a part of the watch face.

![design](/content/watch-face-design.png)

## Environment and SDK setup for development

1. Go to the [Garmin SDK download page](https://developer.garmin.com/connect-iq/sdk/)
1. Follow the steps given on the page to download and install the SDK:
    * Download the SDK manager
    * Launch the downloaded SDK manager
    * Complete first-time setup
    * Within the SDK manager, download the latest Connect IQ SDK and choose the devices you want to develop for (e.g. Fenix 6)
    * Once download finishes, click "Yes" when prompted to use the new SDK version as your active SDK
    * Close the SDK manager
1. Open your Visual Studio Code
1. Follow the steps given on the SDK download page for **installing the Visual Studio Code Monkey C Extension**:
    * Go to Extensions (```cmd + shift + X``` on Mac)
    * Search for "Monkey C", select the one from Garmin ([direct link](https://marketplace.visualstudio.com/items?itemName=garmin.monkey-c))
    * Install the extension
    * Restart VSCode
    * Go to command palette with ```ctrl + shift + p``` (```cmd + shift + p``` on mac)
    * Run "**Verify installation**" under Monkey C: Verify Installation

## Starting a project

## Code structure and what goes where

## Best practices to consider before starting development

### Device aspect ratio and round vs. square watches

Know what you are developing for and choose your build targets accordingly. The watch face in this guide is clearly specifically designed for a round watch face (with a 1:1 aspect ratio - i.e. a circle), so making it available on square-faced watches would lead to a pretty bad experience.

Note that Garmin also has some stranger devices with non-square aspect ratios, if you plan on making your app available on *every* device, you should then correspondingly test it in the simulator for each aspect ratio and screen shape.

### Scaling elements based on resolution

Devices that have the same aspect ratio and screen shape may still have significantly different resolutions. Resolution tends to increase over time with each new model, and resolutions are generally much higher with more smartwatch-oriented devices. For example, the screen resolution of the Fenix 7S is 240x240, while the resolution of the Venu is 390x390.

This means that when you define the position of an element on the screen, you should always define it in a **relative** way. If an element is drawn "10 pixels to the right, from the left edge of the screen", the gap will look large on a Fenix but tiny on a Venu, so the proportions of your app are distorted. Similarly for the thickness of an element - a line 2 pixels in thickness will be decently legible on an older Fenix, but will look oddly small on a Venu.

To preserve the proportions of your design, elements have to be both **positioned** and **scaled** depending on the resolution of the screen.

## Rendering time in digital format

* Time formatting reference
* Styling

## Creating data 'gauges'

## Compiling the project

## Uploading to Garmin Connect IQ store

## Downloading the watch face to your watch

## References

* [Garmin Connect IQ SDK](https://developer.garmin.com/connect-iq/sdk/)
* [Monkey C VS Code extension](https://marketplace.visualstudio.com/items?itemName=garmin.monkey-c)
