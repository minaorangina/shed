require('babel-register')({
  presets: ['es2015']
});
const express = require('express');
const morgan = require('morgan');
const router = require('./router');

const port = process.env.PORT || 3008;

const app = express();
app.use(morgan('combined'));
app.use(express.static('public'));

app.listen(port, () => console.info(`🌍 Server is listening on ${port}`));
