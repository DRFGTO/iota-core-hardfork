import * as React from 'react';
import {inject, observer} from "mobx-react";
import {ExplorerStore} from "../stores/ExplorerStore";
import {Transaction} from "./Transaction";
import {ExplorerTransactionMetadata} from "./ExplorerTransactionMetadata";
import Container from "react-bootstrap/Container";

interface Props {
    explorerStore?: ExplorerStore;
}

@inject("explorerStore")
@observer
export class TransactionPayload extends React.Component<Props, any> {
    render() {
        let {payload} = this.props.explorerStore;
        let txID = payload.txID;
        let tx = payload.transaction;

        return (
            <Container>
                <ExplorerTransactionMetadata txId={txID}/>
                <Transaction txID={txID} tx={tx}/>
            </Container>
        );
    }
}
