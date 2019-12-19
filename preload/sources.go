package preload

// EFF-maintained STARTTLS Everywhere preload list.
// https://starttls-everywhere.org/
var STARTTLSEverywhere = Source{
	ListURI: "https://dl.eff.org/starttls-everywhere/policy.json",
	SigURI:  "https://dl.eff.org/starttls-everywhere/policy.json.asc",

	/*
	   pub   rsa2048 2018-04-18 [SC]
	     B693F33372E965D76D55368616EEA65D03326C9D
	   uid           STARTTLS Everywhere Team <starttls-policy@eff.org>
	   sub   rsa2048 2018-04-18 [S]

	   Key is stripped the to bare minimum needed for signature verification. E.g.
	   encryption and authentication subkeys are removed.
	*/
	SigKey: `-----BEGIN PGP PUBLIC KEY BLOCK-----

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
S/SBhVRBN8L4SJ3ywKB86jnDalI=
=InKu
-----END PGP PUBLIC KEY BLOCK-----`,
}
