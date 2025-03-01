# Spotify Tmux Integration

This application integrates with the Spotify API to provide various features. After setting up your **Spotify API credentials**, you can run the pre-built binary on your machine.

## Prerequisites

Before you can run the application, you'll need to obtain your **Spotify API credentials** (Client ID and Client Secret).

### Step 1: Obtain Your Spotify API Credentials

1. Go to the [Spotify Developer Dashboard](https://developer.spotify.com/dashboard/applications).
2. Log in with your Spotify account or create one if you don't have one.
3. Click on **Create an App**.
4. Fill in the required fields (App name, description, etc.).
5. Once your app is created, you will be provided with a **Client ID** and **Client Secret**.

   > **Important:** Keep your **Client Secret** confidential. It should not be exposed in public repositories.

### Step 2: Set Up the `.env` File

1. In the root directory of the application, create a `.env` file.
2. Add the following content, replacing `your-client-id` and `your-client-secret` with the values you obtained from the Spotify Developer Dashboard:

CLIENT_ID=your-client-id CLIENT_SECRET=your-client-secret (wrap them in double quotes)


3. **Important:** Ensure that you **never commit your `.env` file** to any public repository. You can add it to `.gitignore` to prevent this:

.env


### Step 3: Download and Run the Pre-Built Binary

Once you've set up the `.env` file, you can run the pre-built binary for your operating system:

#### For Linux & MacOS:

```bash
chmod +x spotify-tmux
Run the application:

./spotify-tmux
```
