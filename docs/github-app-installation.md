# GitHub App Installation Guide

This guide walks you through installing your GitHub App on the Kinetik repositories.

## Step 1: Access Your GitHub App

After creating your GitHub App, you need to install it on your repositories.

1. Go to your GitHub App settings page:
   - URL format: `https://github.com/settings/apps/[your-app-name]`
   - Or go to: Settings → Developer settings → GitHub Apps → Click your app name

## Step 2: Click "Install App"

1. In the left sidebar of your app settings page, look for **"Install App"**
2. Click on it

## Step 3: Choose Where to Install

You'll see a page titled "Install [Your App Name]"

1. You'll see your account/organization listed
2. Click the green **"Install"** button next to your account name

## Step 4: Select Repositories

On the installation configuration page:

1. **Repository access**: Choose **"Only select repositories"**

   ⚠️ Do NOT choose "All repositories" - we only want specific repos

2. Click the **"Select repositories"** dropdown

3. Search for and select these three repositories:
   - ✅ `TomasBack2Future/Kinetik`
   - ✅ `TomasBack2Future/KinetikServer`
   - ✅ `TomasBack2Future/kinetik_agent`

4. Click the green **"Install"** button at the bottom

## Step 5: Verify Installation

### Method 1: Check via Repository Settings

For each repository:

1. Go to the repository page (e.g., `https://github.com/TomasBack2Future/Kinetik`)
2. Click **Settings** tab
3. In left sidebar, click **Integrations** → **GitHub Apps**
4. You should see your app "Kinetik Automation Bot" listed
5. Status should show as "Installed"

### Method 2: Check via App Settings

1. Go back to your app settings: Settings → Developer settings → GitHub Apps → Your app
2. Click **"Install App"** in the left sidebar
3. You should see "Installed" next to your account
4. Click "Configure" to see which repos are selected

## Step 6: Update Webhook URL (Later)

After you deploy your webhook server:

1. Go to your app settings
2. In the **"Webhook"** section
3. Update **"Webhook URL"** to:
   - Local testing: `https://[your-ngrok-url].ngrok.io/github/webhook`
   - Production: `https://your-domain.com/github/webhook`
4. Click **"Save changes"**

## Troubleshooting

### App Not Showing in Repository Settings

- **Problem**: App doesn't appear in repository Integrations
- **Solution**:
  1. Check you installed it on the correct repositories
  2. Go to app settings → Install App → Configure
  3. Verify the repositories are selected
  4. You might need to uninstall and reinstall

### Can't Install App

- **Problem**: "Install" button is grayed out
- **Solution**:
  1. Make sure you saved the app after creating it
  2. Make sure you have admin access to the repositories
  3. Try refreshing the page

### Wrong Repositories Selected

- **Problem**: Installed on wrong repos or all repos
- **Solution**:
  1. Go to app settings → Install App → Configure
  2. Change repository access selection
  3. Click "Save"

## What Happens After Installation?

Once installed, your app will:

1. **Receive webhooks** from the selected repositories for:
   - New issues
   - Issue comments
   - Pull requests
   - PR reviews and comments

2. **Have permissions** to:
   - Read repository contents
   - Read and write issues
   - Read and write pull requests
   - Read metadata

3. **Be visible** in:
   - Repository settings under Integrations
   - Issue/PR sidebars (depending on configuration)

## Testing the Installation

Before deploying your webhook server, you can test if webhooks are being sent:

1. Go to app settings → **"Advanced"** tab
2. Scroll down to **"Recent Deliveries"**
3. Create a test issue in one of your repositories
4. Refresh the Recent Deliveries page
5. You should see a webhook delivery attempt (it will fail until your server is running)

## Next Steps

After successful installation:

1. ✅ App is installed on all three repositories
2. ✅ Webhook secret is saved
3. ⏭️ Set up database
4. ⏭️ Configure environment variables
5. ⏭️ Deploy webhook server
6. ⏭️ Update webhook URL in app settings
7. ⏭️ Test end-to-end workflow

## Quick Reference

**App Settings URL**: `https://github.com/settings/apps/[your-app-name]`

**Install App**: App Settings → Install App (left sidebar) → Install button

**Configure Repos**: App Settings → Install App → Configure

**Verify Installation**: Repository → Settings → Integrations → GitHub Apps

**Webhook URL**: App Settings → General → Webhook section
