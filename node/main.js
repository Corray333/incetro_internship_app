const express = require('express');
const { Client } = require('@notionhq/client');
const { NotionToMarkdown } = require('notion-to-md');
const bodyParser = require('body-parser');

const app = express();
const port = 3002;

// Notion API setup
const notion = new Client({ auth: 'secret_mNX6HmRkgT9UxplFxJdW7ydRQCiLZWBvZ80o2QS11Ca' });
const n2m = new NotionToMarkdown({ notionClient: notion });

app.use(bodyParser.json());

app.get('/get-md', async (req, res) => {
  try {
    const { pageId } = req.query;
    if (!pageId) {
      return res.status(400).send('Page ID is required');
    }

    // Retrieve Notion page content
    const mdblocks = await n2m.pageToMarkdown(pageId);
    const mdString = n2m.toMarkdownString(mdblocks);

    res.setHeader('Content-Disposition', 'attachment; filename=page.md');
    res.setHeader('Content-Type', 'application/json');
    if (mdString.parent != undefined){
      res.send({
        md: mdString.parent
      });
    } else {
      res.send({
        md: ""
      })
    }

  } catch (error) {
    console.error(error);
    res.status(500).send('Error retrieving Notion page content');
  }
});

app.listen(port, () => {
  console.log(`Server is running on http://localhost:${port}`);
});
