/*
 * This file (which will be your service worker)
 * is picked up by the build system ONLY if
 * quasar.conf > pwa > workboxPluginMode is set to "InjectManifest"
 */

import { precacheAndRoute } from 'workbox-precaching';

// Use with precache injection
// eslint-disable-next-line no-restricted-globals,no-underscore-dangle
precacheAndRoute(self.__WB_MANIFEST);
