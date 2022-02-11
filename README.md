# bachelor-projekt
Peer-To-Peer Netzwerk zum erzeugen von Signaturen mithilfe des FROST Signaturverfahrens

## Projekt Setup
Als erstes muss Golang installiert werden. Eine Anleitung, wie Golang installiert werden muss und wie die Umgebungsvariablen gesetzt werden müssen, kann [hier](https://go.dev/doc/install)  gefunden werden.

Als nächstes muss das Repository runtergeladen werden.
In einem Terminalfenster:
`git clone https://github.com/BrendanMeins/bachelor-projekt.git`

Anschließend in das Verzeichnis wechseln:
`cd bachelor-projekt`

Nun müssen die Abhängigkeiten installiert werden.
`go get`

Anschließend kann der Peer gestartet werden durch den Befehl:
`go run . `

Sollte eine Fehlermeldung erscheinen, weil ein Paket nicht installiert wurde, dann kann das Paket nachträglich nochmal installiert werden.
`go get <name>`

## Mehrere Peers starten

Öffne das Projekt `bachelor-projekt` in mehreren Terminals, in jedem Terminal muss der Peer mit `go run . ` gestartet werden. Die Peers finden sich automatisch. 
In den Ausgaben ist zu sehen, was die Peers gerade machen. Zum Beispiel gibt es eine Ausgabe, wenn der Peer einen anderen Peer findet. 
Wenn alle Peers gestartet werden, kann durch den Befehl `keygen` die Schlüsselerzeugung eingeleitet werden. Die Terminal Ausgabe enthält die Nachrichten, die Verschickt werden. Zu letzt wird der gemeinsame öffentliche Signaturschlüssel ausgegeben.

## Eine Signatur erzeugen
Nachdem ein Schlüssel erzeugt wurde, kann durch den Befehl `sign <nachricht> ` eine Nachricht signiert werden. Die Terminal Ausgabe gibt die Verschickten Nachrichten sowie die Signatur zurück.