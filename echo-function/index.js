exports.echo = (req, res) => {
  console.log(`Handling HTTP event ${req.body.eventID}`)

  var json = JSON.stringify({ 
    body: req.body.data.body,
    statusCode: 200, 
  });

  res.status(200).send(json);
};
