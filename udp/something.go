package udp

//
//func foo() {
//	cc := ipv4.NewConn(nconn.(net.Conn))
//	if err != nil {
//		fmt.Printf("error .NewRawConn(): %s\n", err)
//		return
//	}
//	for {
//		fmt.Println("Entering main loop")
//		buf := make([]byte, 1024)
//		_, _, _, err = pconn.ReadFrom(buf)
//		if err != nil {
//			fmt.Printf("error pconn.ReadFrom(): %s\n", err)
//			continue
//		}
//		m := ipv4.Message{
//			Buffers: [][]byte{
//				make([]byte, 1024),
//			},
//			OOB: ipv4.NewControlMessage(ipv4.FlagTTL | ipv4.FlagSrc | ipv4.FlagDst | ipv4.FlagInterface),
//		}
//		if err := cc.RecvMsg(&m, 0); err != nil {
//			fmt.Printf("error RecvMsg(): %s\n", err)
//			continue
//		}
//		payload := m.Buffers[0][:m.N]
//		fmt.Printf("OOB: %v\n", m.OOB)
//		fmt.Printf("Addr: %s\n", m.Addr)
//		fmt.Printf("Message: %v\n", hex.EncodeToString(payload))
//
//		//out := &ipv4.Message{
//		//	Buffers: [][]byte{payload},
//		//}
//	}
//
//}
