# Demo walkthrough

[Download demo (MP4)](https://github.com/chuhezi/socialai/raw/refs/heads/main/docs/demo/socialai-demo.mp4)

GitHub does not preview this video on the repository file page. Download it and open it in your video player.

The public recording is approximately 3 minutes 36 seconds and demonstrates:

1. The sign-in screen and authenticated creation view.
2. An actual image-generation request, preview, and publication.
3. The multi-column gallery, image viewer, reactions, and deletion.
4. Uploading an original photograph.
5. Uploading and playing a short video, including fullscreen playback.
6. Keyword search, user search, and the personal-post filter.
7. Signing out.

The current recording does not clearly demonstrate caption editing across two simultaneous views or responsive window resizing. Those features should be shown separately when discussing them. Showing a login screen alone is not evidence for the underlying JWT/session checks; the backend tests and implementation provide that evidence.

The original 3:49 recording is preserved locally. The public version omits original seconds 8–21 to remove login setup and a browser autofill suggestion; the remaining sequence is preserved without speeding up generation. Its silent audio track is omitted.

The demo runs the React frontend locally against a deployed Go API. A localhost URL is not a public live-demo URL. No performance or user-growth statistics are inferred from the recording.
