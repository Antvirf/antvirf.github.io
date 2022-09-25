+++ 
date = 2022-09-25
title = "Turning a static Hugo site to a Progressive Web Application"
description = "Documenting the key steps required to turn a static Hugo site into a Progressive Web Application."
author = "Antti Viitala"
tags = [
    "devops",
    "development"
]
+++

After learning the basics of [Progressive Web Applications (PWAs)](https://web.dev/progressive-web-apps/) at work, I decided to update this site to fill the requirements of a PWA. The commit with those changes can be found [here](https://github.com/Antvirf/antvirf.github.io/commit/01377e738439aedf72c4fd676f24234e31cfe7d9). This post contains the basic details required to turn a static Hugo site like this one into a PWA.

## 'Additional' requirements for a PWA

Beyond what is listed here, a PWA has other requirements a well - for example the site must be served over HTTPS - but these are the 'additional' steps I had to take to enable the functionality on this site.

* Broad set of icons for different platforms
* A Manifest file ```site.webmanifest``` or ```manifest.json``` - ([ref](https://web.dev/learn/pwa/web-app-manifest/))
* A JavaScript "service worker" - ([ref](https://web.dev/learn/pwa/service-workers/))
* Some added HTML to refer to the various icons and provide other meta data


## Creating icons

A PWA will need various sizes of icons for use with different platforms. If you already have a base image, you can feed it through [favicomatic](https://favicomatic.com/) to generate all the different variations. If you don't already have a base icon, one good resource is this [launcher icon generator](https://romannurik.github.io/AndroidAssetStudio/icons-launcher.html).

These icons need to be served along with your site, from the ```static``` folder.

## Updating / creating the apps web manifest

The name of the manifest file seems to vary, but my site was using ```site.webmanifest```. The format should follow JSON, and the file should be provided in the ```static``` folder of your site. 

The ```site.manifest``` of this site is included below and can serve as a starting point.

```json
{
  "name": "AVIITALA.com",
  "short_name": "AVIITALA.com",
  "theme_color": "#212121",
  "background_color": "#212121",
  "display": "standalone",
  "start_url": "/",
  "orientation": "portrait",
  "icons": [
    {
      "src": "/images/favicon-128.png",
      "sizes": "128x128",
      "type": "image/png"
    },
    {
      "src": "/images/apple-touch-icon-144x144.png",
      "sizes": "144x144",
      "type": "image/png"
    },
    {
      "src": "/images/apple-touch-icon-152x152.png",
      "sizes": "152x152",
      "type": "image/png"
    },
    {
      "src": "/images/favicon-196x196.png",
      "sizes": "196x196",
      "type": "image/png"
    },
    {
      "src": "/images/splash.png",
      "sizes": "512x512",
      "type": "image/png",
      "purpose": "maskable"
    }
  ]
}  
```

## Creating a minimal service worker

For this site, I didn't need any 'real' functionality from a service worker, and hence just added this minimal service worker to fill the requirement:

```javascript
self.addEventListener ('fetch', function(event) {
});
```

## Updates to the HTML

The HTML code has to contain certain meta tags with information for PWA usage. The below sample from this page omits listing every icon but contains all other essentials.

```html
<meta name="mobile-web-app-capable" content="yes" />
<meta name="apple-mobile-web-app-title" content="AVIITALA.com" />
<meta name="apple-mobile-web-app-status-bar-style" content="black-translucent" />
<meta name="application-name" content="Antti Viitala"/>

<!-- various other "apple touch-icon-precomposed" lines -->
<link rel="apple-touch-icon-precomposed" sizes="57x57" href="https://aviitala.com/images/apple-touch-icon-57x57.png" />

<!-- various other "icon" lines -->
<link rel="icon" type="image/png" href="https://aviitala.com/images/favicon-196x196.png" sizes="196x196" />

<!-- various other "msapplication" related lines -->
<meta name="msapplication-TileColor" content="#FFA500" />
<meta name="msapplication-TileImage" content="https://aviitala.com/images/mstile-144x144.png" />

<!-- load the minimal service worker -->
<script defer>
if ('serviceWorker' in navigator) {
navigator.serviceWorker.register('/sw.js');
};
</script>
```

## Testing the app

To ensure things are working properly, [Lighthouse](https://chrome.google.com/webstore/detail/lighthouse/blipmdconlkpinefehnmjammfjpmpbjk?hl=en) is a great extension for Google Chrome to ['audit'](https://web.dev/lighthouse-pwa/) your site for crucial PWA requirements.

