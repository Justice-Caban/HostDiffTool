#!/usr/bin/env node

/**
 * Test script for the new git-diff style DiffViewer component
 *
 * This script:
 * 1. Uploads two snapshots with differences
 * 2. Compares them using the web UI
 * 3. Verifies the git-diff visualization is displayed
 * 4. Takes a screenshot of the diff viewer
 */

const puppeteer = require('puppeteer');
const path = require('path');
const fs = require('fs');

const BASE_URL = 'http://localhost';

async function sleep(ms) {
  return new Promise(resolve => setTimeout(resolve, ms));
}

async function runTest() {
  console.log('🚀 Starting DiffViewer E2E Test...\n');

  const browser = await puppeteer.launch({
    headless: true,
    args: ['--no-sandbox', '--disable-setuid-sandbox']
  });

  const page = await browser.newPage();
  await page.setViewport({ width: 1920, height: 1080 });

  try {
    // Navigate to the application
    console.log('📍 Navigating to application...');
    await page.goto(BASE_URL, { waitUntil: 'networkidle2' });
    await sleep(1000);

    // Upload first snapshot (or skip if it already exists)
    console.log('📤 Uploading first snapshot...');
    const file1Path = path.resolve(__dirname, 'assets/host_snapshots/host_125.199.235.74_2025-09-10T03-00-00Z.json');
    const fileInput = await page.$('input[type="file"]');
    await fileInput.uploadFile(file1Path);
    await sleep(2000);

    let resultText = await page.$eval('.card pre', el => el.textContent);
    console.log(`   Result: ${resultText.trim()}`);

    if (resultText.includes('Snapshot uploaded')) {
      console.log('   ✅ First snapshot uploaded successfully\n');
    } else if (resultText.includes('UNIQUE constraint failed')) {
      console.log('   ⚠️  First snapshot already exists (skipping)\n');
    } else {
      console.log('   ⚠️  Unexpected response, but continuing...\n');
    }

    // Upload second snapshot (or skip if it already exists)
    console.log('📤 Uploading second snapshot...');
    const file2Path = path.resolve(__dirname, 'assets/host_snapshots/host_125.199.235.74_2025-09-15T08-49-45Z.json');
    await fileInput.uploadFile(file2Path);
    await sleep(2000);

    resultText = await page.$eval('.card pre', el => el.textContent);
    console.log(`   Result: ${resultText.trim()}`);

    if (resultText.includes('Snapshot uploaded')) {
      console.log('   ✅ Second snapshot uploaded successfully\n');
    } else if (resultText.includes('UNIQUE constraint failed')) {
      console.log('   ⚠️  Second snapshot already exists (skipping)\n');
    } else {
      console.log('   ⚠️  Unexpected response, but continuing...\n');
    }

    // Get host history
    console.log('📜 Getting host history...');
    await page.type('input[type="text"]', '125.199.235.74');
    await sleep(500);
    await page.click('button:nth-of-type(1)');
    await sleep(2000);

    // Check if snapshots are displayed
    const snapshots = await page.$$('.snapshot-item');
    console.log(`   Found ${snapshots.length} snapshots`);

    if (snapshots.length < 2) {
      throw new Error('Not enough snapshots found for comparison');
    }
    console.log('   ✅ Host history retrieved successfully\n');

    // Select two snapshots
    console.log('🎯 Selecting two snapshots for comparison...');
    await snapshots[0].click();
    await sleep(500);
    await snapshots[1].click();
    await sleep(500);

    // Check if both are selected
    const selectedSnapshots = await page.$$('.snapshot-item.selected');
    console.log(`   Selected ${selectedSnapshots.length} snapshots`);

    if (selectedSnapshots.length !== 2) {
      throw new Error('Failed to select two snapshots');
    }
    console.log('   ✅ Two snapshots selected\n');

    // Click compare button
    console.log('🔍 Comparing snapshots...');
    // Wait for compare button to appear (it only shows when 2 snapshots are selected)
    await page.waitForSelector('.history-list button', { timeout: 5000 });
    const compareButton = await page.$('.history-list button');
    if (!compareButton) {
      throw new Error('Compare button not found');
    }
    await compareButton.click();
    await sleep(3000);

    // Check if DiffViewer is displayed
    console.log('🎨 Verifying git-diff visualization...');
    const diffViewer = await page.$('.diff-viewer');

    if (!diffViewer) {
      console.log('   ❌ DiffViewer component not found!');

      // Check if old format is still showing
      const preElement = await page.$('.card pre');
      if (preElement) {
        const preText = await page.evaluate(el => el.textContent, preElement);
        console.log('   Old format detected:', preText.substring(0, 100) + '...');
      }

      throw new Error('DiffViewer component not rendered');
    }

    console.log('   ✅ DiffViewer component found!\n');

    // Verify git-diff style elements
    console.log('🔎 Checking git-diff style elements...');

    const diffHeader = await page.$('.diff-header');
    if (!diffHeader) {
      throw new Error('Diff header not found');
    }
    console.log('   ✅ Diff header present');

    const diffFileHeader = await page.$('.diff-file-header');
    if (!diffFileHeader) {
      throw new Error('Diff file header not found');
    }
    const fileHeaderText = await page.evaluate(el => el.textContent, diffFileHeader);
    console.log(`   ✅ File header: "${fileHeaderText}"`);

    const hunkHeaders = await page.$$('.diff-hunk-header');
    console.log(`   ✅ Found ${hunkHeaders.length} hunk header(s)`);

    const addedLines = await page.$$('.diff-line-added');
    console.log(`   ✅ Found ${addedLines.length} added line(s) (green)`);

    const removedLines = await page.$$('.diff-line-removed');
    console.log(`   ✅ Found ${removedLines.length} removed line(s) (red)`);

    const changedLines = await page.$$('.diff-line-context');
    console.log(`   ✅ Found ${changedLines.length} context line(s)`);

    const diffFooter = await page.$('.diff-footer');
    if (!diffFooter) {
      throw new Error('Diff footer not found');
    }
    const footerText = await page.evaluate(el => el.textContent, diffFooter);
    console.log(`   ✅ Footer stats: "${footerText.trim()}"`);

    // Check for specific diff content
    console.log('\n📊 Verifying diff content...');

    // Check for added services section
    const addedServicesHeader = await page.evaluate(() => {
      const headers = Array.from(document.querySelectorAll('.diff-hunk-header'));
      return headers.find(h => h.textContent.includes('Added Services'));
    });

    if (addedServicesHeader) {
      console.log('   ✅ "Added Services" section found');
    } else {
      console.log('   ⚠️  "Added Services" section not found (may be no additions)');
    }

    // Check for changed services section
    const changedServicesHeader = await page.evaluate(() => {
      const headers = Array.from(document.querySelectorAll('.diff-hunk-header'));
      return headers.find(h => h.textContent.includes('Modified Services'));
    });

    if (changedServicesHeader) {
      console.log('   ✅ "Modified Services" section found');
    } else {
      console.log('   ⚠️  "Modified Services" section not found (may be no changes)');
    }

    // Check for CVE sections
    const cveHeaders = (await page.evaluate(() => {
      const headers = Array.from(document.querySelectorAll('.diff-hunk-header'));
      return headers.filter(h => h.textContent.includes('Vulnerabilities'));
    })) || [];

    if (cveHeaders.length > 0) {
      console.log(`   ✅ Found ${cveHeaders.length} CVE section(s)`);
    } else {
      console.log('   ⚠️  No CVE sections found (may be no vulnerability changes)');
    }

    // Take a screenshot
    console.log('\n📸 Taking screenshot...');
    await page.screenshot({
      path: 'diff_viewer_screenshot.png',
      fullPage: true
    });
    console.log('   ✅ Screenshot saved to: diff_viewer_screenshot.png\n');

    // Check color styling
    console.log('🎨 Verifying color styling...');
    const addedLineStyle = await page.evaluate(() => {
      const addedLine = document.querySelector('.diff-line-added');
      if (!addedLine) return null;
      const style = window.getComputedStyle(addedLine);
      return {
        backgroundColor: style.backgroundColor,
        borderLeftColor: style.borderLeftColor
      };
    });

    if (addedLineStyle) {
      console.log(`   ✅ Added line styling: bg=${addedLineStyle.backgroundColor}, border=${addedLineStyle.borderLeftColor}`);
    }

    const removedLineStyle = await page.evaluate(() => {
      const removedLine = document.querySelector('.diff-line-removed');
      if (!removedLine) return null;
      const style = window.getComputedStyle(removedLine);
      return {
        backgroundColor: style.backgroundColor,
        borderLeftColor: style.borderLeftColor
      };
    });

    if (removedLineStyle) {
      console.log(`   ✅ Removed line styling: bg=${removedLineStyle.backgroundColor}, border=${removedLineStyle.borderLeftColor}`);
    }

    console.log('\n✨ All tests passed! Git-diff visualization is working correctly.\n');
    console.log('🎉 SUCCESS: DiffViewer component is displaying snapshots in git-diff style!\n');

  } catch (error) {
    console.error('\n❌ Test failed:', error.message);
    console.error(error.stack);

    // Take error screenshot
    await page.screenshot({
      path: 'diff_viewer_error_screenshot.png',
      fullPage: true
    });
    console.error('Error screenshot saved to: diff_viewer_error_screenshot.png');

    process.exit(1);
  } finally {
    await browser.close();
  }
}

runTest().catch(error => {
  console.error('Fatal error:', error);
  process.exit(1);
});
