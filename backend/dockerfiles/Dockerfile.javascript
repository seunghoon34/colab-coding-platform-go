FROM node:14-alpine
WORKDIR /app
COPY script.js .
CMD ["node", "script.js"]