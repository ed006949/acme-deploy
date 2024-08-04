package io_ssh

// import (
// 	"crypto"
// 	"encoding/base64"
// 	"encoding/pem"
//
// 	"golang.org/x/crypto/ed25519"
// 	"golang.org/x/crypto/ssh"
// )

// func sshKeygenTest() {
// 	pub, priv, err := ed25519.GenerateKey(nil)
// 	if err != nil {
// 		panic(err)
// 	}
// 	p, err := ssh.MarshalPrivateKey(crypto.PrivateKey(priv), "")
// 	if err != nil {
// 		panic(err)
// 	}
// 	privateKeyPem := pem.EncodeToMemory(p)
// 	privateKeyString := string(privateKeyPem)
// 	publicKey, err := ssh.NewPublicKey(pub)
// 	if err != nil {
// 		panic(err)
// 	}
// 	publicKeyString := "ssh-ed25519" + " " + base64.StdEncoding.EncodeToString(publicKey.Marshal()) + " " + ""
// 	Printf("Private Key:\n%s\n", privateKeyString)
// 	Printf("Public Key:\n%s\n", publicKeyString)
// }
