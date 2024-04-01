const targetUrl = 'http://localhost:8000';

const PROXY_CONFIG = [
  {
    context: ["/api"],
    target: targetUrl,
    secure: false,
    changeOrigin: true,
    pathRewrite: { "^/api": "" }
  }
];

module.exports = PROXY_CONFIG;