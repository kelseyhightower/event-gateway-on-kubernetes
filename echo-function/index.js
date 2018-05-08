exports.echo = (req, res) => {
  console.log(`Handling HTTP event ${req.body.eventID}`)

  var headers = new Object();
  headers['Compute-Type'] = 'function';

  var json = JSON.stringify({
    body: req.body.data.body,
    headers: headers,
    statusCode: 200, 
  });

  res.status(200).send(json);
};
