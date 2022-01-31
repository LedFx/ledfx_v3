package color

import "fmt"

func gradientBodyBuilder(css string) []byte {
	return []byte(fmt.Sprintf(`<html>
   <head>
      <meta charset="utf-8">
      <style>
         html, body {
         height: 100%%;
         margin: 0;
         overflow: hidden;
         }
         /* Items inside body will be centered vertically and horizontally */
         body {
         display: flex;
         justify-content: center;
         align-items: center;
         }
         .test-element {
         width: 1500vw;
         height: 70vh;
         }
      </style>
   </head>
   <body>
      <div class="test-element" style="background-image: %s;"></div>`, css))
}
