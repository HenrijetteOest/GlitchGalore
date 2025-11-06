Implement Ricart-Argawala

R1(Spec)
- vi skal have nodes 
        dvs: én fil med client-server kode i samme fil/kode
- a Critical Section 
    dvs: en function eller variable som der skal Lock rundt om
    a Critical section represents a sensitive system operation
    skal være enden
        by print statements
        eller
        writing to  shared database on the network
- any node(ONE) may request access to the critical section at any time

R2(Safety)
- Only ONE node may enter the Critical Section at any time

R3(Liveliness)
- any node that request access wil EVENTUALLY Gain access.

