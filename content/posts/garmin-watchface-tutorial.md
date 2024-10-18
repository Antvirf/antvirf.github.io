+++
author = "Antti Viitala"
title = "Making a custom Garmin watch face"
date = "2022-10-06"
description = "Use Visual Studio Code and the Garmin SDK to build a custom digital watch face and release it to the Garmin ConnectIQ store."
tags = [
    "development"
]
thumbnail = "/content/header.png"
images = ["/content/header.png"]
series = []
+++

![header](/content/header.png)

<!-- 
1. Picture of my shitty drawing
2. Picture in Garmin simulator
3. Picture of Fenix 6 running the watch face

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
1. [Run the Garmin device simulator](#running-the-simulator-for-the-first-time)
1. [Create the digital time display](#rendering-time-in-digital-format)
1. [Create a function to draw 'gauges'](#creating-data-gauges)
1. [Configure data gauges](#creating-data-gauges)
1. [Compile the project](#compiling-the-project)
1. [Upload to Garmin Connect IQ store](#uploading-to-garmin-connect-iq-store)
1. [Download the watch face to your watch](#downloading-the-watch-face-to-your-watch)

The repository for this guide can be found [here](https://github.com/Antvirf/garmin-watch-face-guide).

## Design

The design for this watch face is very simple and prioritizes legibility of a small number of key metrics that I care about:

* **Current time**: Digital 24-hour format, hours and minutes only, without separator for hours/minutes - e.g. 1527 for 3:27pm.
* **Current date**: Without year, in dd-mm format - e.g. 31-12 for 31st of December
* **Battery**: Overall battery %
* **Steps**: Show today's steps versus a pre-defined (here hardcoded) target - e.g. 10k
* **Sunrise/sunset**: Sadly Garmin SDK does not provide these values for custom watch faces, so the values are hardcoded to 0615 and 1830.

Based on the above, I created a rough sketch of what to develop, shown below. I added the 'hour' indices to help with placement of the gauges, they will not be rendered as a part of the watch face.

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

1. Open Visual Studio Code
1. Open command palette with ```ctrl + shift + p``` (```cmd + shift + p``` on mac)
1. Type in "Monkey C: New project" and hit enter to start the create project wizard
1. Choose a name for your project - here "mattermetrics"
1. Choose a project type - **"Watch face"**
1. Choose watch face type - **"Simple with settings"**
1. Choose minimum supported API level - depends on what devices you want to build for, here we will go for **3.0.0** to support some older devices as well
1. Hit enter, and you will be asked to choose a folder where to create the project. Note that the project will create it's own folder, so if you choose ```/Documents```, your project will be created at ```/Documents/projectname/```, depending on the project name that you chose.

After these steps, you should have a file structure resembling the below. The next section describes the project and code structure in further detail.

![project structure](/content/project-resources.png)

## Code structure and what goes where

Please note that I've written the below based on my experience - it is pragmatic and works, but may not be fully correct. If you are keen to go in-depth with the Connect IQ SDK, I recommend starting with Garmin's resources: [SDK basics](https://developer.garmin.com/connect-iq/connect-iq-basics/), [Core topics](https://developer.garmin.com/connect-iq/core-topics/) and [FAQ](https://developer.garmin.com/connect-iq/connect-iq-faq/).

The ```manifest.xml``` file contains high-level details about your application. While you can edit it directly, the VS Code extension makes this easy - type "Monkey c: edit" in the command palette to see which options can be edited, and use the relevant wizard to change an option.

### ```resources``` - non-code assets and configuration files

The four subdirectories under ```resources``` contain non-code assets and configuration files for your watch face application. ```properties.xml```, ```settings.xml``` and ```strings.xml``` together define the user-configurable options that you may want to implement - for example choosing between different date formats or choosing different data points for the watch face. ```properties.xml``` contains the default values, ```settings.xml``` describes the type of setting and options available to the user, and ```strings.xml``` maps settings options to machine-readable variable values.

By default, the template comes with three pre-defined settings:

* Background color: Color for watch face background, defaults to black
* Foreground color: Color for rendering time, defaults to red
* Use military time: True/false for whether to use this format, defaults to false

```layouts``` can be used to define simpler and static elements of your watch face, for example drawing a bitmap logo or rendering time as text. This file will be updated later to customize how time and date is displayed.

The ```drawables``` directory contains your app's launcher icon. This may be more relevant for other Connect IQ apps, but for watch faces I haven't come across a need to change it.

### ```source``` - your application code

The ```source``` directory contains the application code you will write. Creating a project as described earlier should result in three files being created in this directory:

1. ```projectnameView.mc```: Details of the watch face - **our code will go here!**
1. ```projectnameApp.mc```: The high-level code for a watch face Connect IQ app - you don't need to edit this at all.
1. ```projectnameBackground.mc```: Sets the background on which the ```View``` is drawn - you don't need to edit this at all.

Note that the naming of these files depends of the project name you chose earlier, so in my example these are ```mattermetricsApp.mc```, ```mattermetricsView.mc```, ```mattermetricsBackground.mc```.

### The ```projectnameView.mc``` file

This will be the most important file for our watch face code. The contents of the file are well explained by comments included as part of the template - abridged version of that file with just the main class and its key functions is shown below:

```c
class mattermetricsView extends WatchUi.WatchFace {

    function initialize() {
        WatchFace.initialize();
    }

    // Load your resources here
    function onLayout(dc as Dc) as Void {}

    // Called when this View is brought to the foreground. Restore
    // the state of this View and prepare it to be shown. This includes
    // loading resources into memory.
    function onShow() as Void {}

    // Update the view
    function onUpdate(dc as Dc) as Void {}

    // Called when this View is removed from the screen. Save the
    // state of this View here. This includes freeing resources from
    // memory.
    function onHide() as Void {}

    // The user has just looked at their watch. Timers and animations may be started here.
    function onExitSleep() as Void {}

    // Terminate any active timers and prepare for slow updates.
    function onEnterSleep() as Void {}
}
```

Primarily we will work within the ```onUpdate``` function - you can assume this is run every minute to update the contents of the watch face.

The other functions, especially dealing with hide/show and entering/exiting sleep, become relevant for more complex watch faces and for further optimization of the power consumption of your code.

## Best practices to consider before starting development

### Device aspect ratio and round vs. square watches

Know what you are developing for and choose your build targets accordingly. The watch face in this guide is clearly specifically designed for a round watch face (with a 1:1 aspect ratio - i.e. a circle), and rendering the gauges for example would break on other types of screens. Note that Garmin also has some stranger devices with non-square aspect ratios, so if you plan on making your app available on *every* device, you should then correspondingly test it in the simulator for each aspect ratio and screen shape.

### Scaling elements based on resolution

Devices that have the same aspect ratio and screen shape may still have significantly different resolutions. For example, the screen resolution of the Fenix 7S is 240x240, while the resolution of the Venu is 390x390.

This means that when you define the position of an element on the screen, you should always define it in a **relative** way. If an element is drawn "10 pixels to the right, from the left edge of the screen", the gap will look large on a Fenix but tiny on a Venu, so the proportions of your app are distorted. Similarly for the thickness of an element - a line 2 pixels in thickness will be decently legible on an older Fenix, but will look oddly small on a Venu. To preserve the proportions of your design, elements have to be both **positioned** and **scaled** depending on the resolution of the screen.

For scaling, in my watch faces I define a scaling variable based on the current device screen width vs. the value Fenix 6 I primarily develop for:

```c
var scaler = dc.getWidth()/260.0;
// 1.0 for Fenix 6
// 390/260 = 1.5 for Venu etc.
```

## Adding a target device and language to your project

In order to run the simulator, your project needs to have a target product (=the specific Garmin device) it is intended for. You can set it in Visual Studio code as follows:

1. Open command palette with ```ctrl + shift + p``` (```cmd + shift + p``` on mac)
1. Type "Monkey C: Edit products"
1. Tick each product you want to support - for this example ```Fenix 6 Pro / 6 Sapphire / 6 Pro Solar / 6 Pro Dual Power```
    * **If your device is not on this list, likely the API version requirement configured earlier was set too high. Use the "Monkey C: Edit Application" and decrease the API version number parameter.**
1. Open the command palette and execute "Monkey C: Edit languages"
1. Choose at least one, for example English

[This commit](https://github.com/Antvirf/garmin-watch-face-guide/commit/ea6d87c6b6ab25717bb4f90971641279e837aca8) contains the initialization of the project and setting the target device.

## Running the simulator for the first time

Without changing any of the code, open the ```projectnameView.mc``` and go to the "Run and Debug" section in VS Code or execute the command palette command "Debug: Start debugging" to launch the simulator. This will show your device with the default starting point watch face like below.

Which device is shown by default will depend on the products added to the ```manifest.xml```.

![simulator-screen](/content/simulator.png)

To change any of the device settings - such as preferred time format - choose the simulator window, use the Settings menu from the top bar as shown below.

![simulator-settings](/content/simulator-settings.png)

## Rendering time in digital format

The default watch face already renders time in digital format. Though the design calls for the time to be displayed in 24-hour format, the template code actually uses the user-defined device-level settings for time 12 vs. 24-hour time format, so we get this optional functionality out of the box.

First, as my focus is on military time, I changed the default value for the relevant setting (```UseMilitaryFormat```). I also changed the default value of ```ForegroundColor``` to white in [this commit](https://github.com/Antvirf/garmin-watch-face-guide/commit/80d22cf7e0e0d237ba4250978dc2ef5a4955d707). Later on I removed the background and foreground color setting customization options in [this commit](https://github.com/Antvirf/garmin-watch-face-guide/commit/1a3add89faa8417bcba2a9452ef24771a03a0e7f).

The size of the text is obviously too small - this is fixed by changing the font in ```resources/layouts/layout.xml``` in [this commit](https://github.com/Antvirf/garmin-watch-face-guide/commit/3387e5fe143b444772017b6244cec42327322b70). The different constants describing fonts are listed [here](https://developer.garmin.com/connect-iq/api-docs/Toybox/Graphics.html) in the SDK docs. In this case, I changed the font value to ```Graphics.FONT_SYSTEM_NUMBER_THAI_HOT```. You can see the change displayed below.

![font-before](/content/font-before.png)![font-after](/content/font-after.png)

## Accessing data points and adding date

To add in the date, we make changes in two places - first, add a layout line to ```layout.xml```, and then add code to the ```onUpdate()``` function in ```projectnameView.mc```.

The ```layout.xml``` file needs an additional label. The positioning is done in relative terms - x-coordinate is still centered, but the y-coordinate is set to 20% to position date above the time. A smaller font is also used.

```xml
<label id="DateLabel" x="center" y="20%" font="Graphics.FONT_TINY" justification="Graphics.TEXT_JUSTIFY_CENTER"/>
```

The code to update the date is simple, though note that the lines must be added before ```View.onUpdate(dc)``` is called. Additionally, a new import is added at the top: ```import Toybox.Time.Gregorian;```. Formatting is done with [```Lang.format()```](https://developer.garmin.com/connect-iq/api-docs/Toybox/Lang.html#format-instance_function). The commit with this change can be found [here](https://github.com/Antvirf/garmin-watch-face-guide/commit/3b248a442eb7792c70a85e1d184f8a17d86a5e4f).

```c
function onUpdate(dc as Dc) as Void {
    ... // Existing code to draw time is omitted

    // Get date info from the Toybox.Time.Gregorian package
    var info = Gregorian.info(Time.now(), Time.FORMAT_SHORT);

    // Format 
    var dateString = Lang.format("$1$-$2$", [info.day, info.month]);

    // Find the drawable we added to our layout.xml
    var dateView = View.findDrawableById("DateLabel") as Text;

    // Set the label color, and text value
    dateView.setColor(getApp().getProperty("ForegroundColor") as Number);
    dateView.setText(dateString);

    // Call the parent onUpdate function to redraw the layout level
    View.onUpdate(dc);
}
```

![after date added](/content/date-added.png)

## Creating data 'gauges'

As per the design, the watch face needs to have three different gauge elements, each based on different data. Since the functionality required to draw a gauge can therefore be abstracted and reused, a function is defined that draws a gauge from given a set of inputs. Each gauge needs the following values as inputs:

1. ```Number```: ```start_hour```: hour index at which the gauge starts
1. ```Number```: ```duration```: 'duration' of the gauge in hours, determining its length on the dial
1. ```Number```: ```direction```: direction of gauge rotation 0=ccw, 1, cw
1. ```Float```: ```start_val```: minimum value of the gauge
1. ```Float```: ```end_val```: maximum value of the gauge
1. ```Float```: ```cur_val```: current value of the gauge
1. ```String```: ```cur_label```: text to display for current value
1. ```String```: ```start_label```: text to display at start of the gauge
1. ```String```: ```end_label```: text to display at end of the gauge

This function is called  ```drawGauge()``` in the code, and calls a separate function ```drawHashMarksAndLabels()``` which in turn renders certain elements of the gauge. The actual implementation uses arrays instead of individual arguments, as the maximum number of function parameters is capped to 10 in MonkeyC.

The detailed development of this function - drawing 2D graphics and the mathematics around it - aren't in the scope of this guide, but the code can be found in [this commit](https://github.com/Antvirf/garmin-watch-face-guide/commit/ce090ac847ad098f23d54dfbad2975c63b0d0a20). If you are interested in learning more, the key functions used from the Garmin SDK graphics libraries are [```drawArc()```](https://developer.garmin.com/connect-iq/api-docs/Toybox/Graphics.html#ArcDirection-module), [```drawText()```](https://developer.garmin.com/connect-iq/api-docs/Toybox/Graphics/Dc.html#drawText-instance_function) and [```fillPolygon()```](https://developer.garmin.com/connect-iq/api-docs/Toybox/Graphics/Dc.html#fillPolygon-instance_function).

![with gauges](/content/with-gauges.png)

## Compiling the project

Note that you must have a developer key defined to do this. You can generate one with the extension by running **"Monkey C: Generate Developer Key"**.

1. Open command palette with ```ctrl + shift + p``` (```cmd + shift + p``` on mac)
1. Type **"Monkey C: Export project"**
1. Choose the export location to save the file to

Once finished, you will have a ```projectname.iq``` file, ready for upload to the Garmin store. If you would like to transfer the file to your watch directly instead, use the ```projectname.prg``` file.

## Uploading to Garmin Connect IQ store

1. Navigate to the [Garmin Developer Dashboard](https://apps.garmin.com/en-US/developer/dashboard)
1. Sign in to your Garmin developer account if you have not yet done so
1. Click "Upload an App"
1. Choose the ```projectname.iq``` file exported in the previous step
1. Go through the process, upload app pictures where requested.
1. If desired, tick the box to mark it as a beta application (to only allow yourself to download it later).
1. Once complete, you will have to wait for approval - see the [developer dashboard](https://apps.garmin.com/en-US/developer/dashboard) to check the status.
1. Once the status changes to **Approved**, you can download the app. This may take up to 3 days.

## Transferring a watch face to your watch (offline)

1. Connect your Garmin device in mass media transfer mode
1. Copy your ```projectname.prg``` file to ```/GARMIN/APPS``` folder
1. Disconnect your watch
1. Edit your watch faces and choose your newly created custom watch face

## Downloading the watch face to your watch

Your app must be approved by Garmin before it can be downloaded. Once the approval process is complete, you can search for it on the ConnectIQ store and download to your device.

The watch face created during this guide, is available [here](https://apps.garmin.com/en-US/apps/38b1b25e-3cf7-4993-9fd9-7ced64eb3564).

![final](/content/hero.jpg)

## References

* [Garmin Connect IQ SDK](https://developer.garmin.com/connect-iq/sdk/)
* [Monkey C VS Code extension](https://marketplace.visualstudio.com/items?itemName=garmin.monkey-c)
